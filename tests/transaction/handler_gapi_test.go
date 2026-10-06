package transaction_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/transaction"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TransactionGapiTestSuite struct {
	suite.Suite
	ts                 *tests.TestSuite
	db                 *gorm.DB
	redisClient        redis.UniversalClient
	chConn             clickhouse.Conn
	grpcServer         *grpc.Server
	conn               *grpc.ClientConn
	commandClient      pb.TransactionCommandServiceClient
	queryClient        pb.TransactionQueryServiceClient
	userRepo           user_repo.UserCommandRepository
	cardRepo           card_repo.Repositories
	saldoRepo          saldo_repo.Repositories
	merchantRepo       merchant_repo.Repositories
	customerCardNumber string
	merchantID         int
	merchantApiKey     string
}

func (s *TransactionGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "saldo", "transaction"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	chOpts, err := clickhouse.ParseDSN(s.ts.CHURL)
	s.Require().NoError(err)
	chConn, err := clickhouse.Open(chOpts)
	s.Require().NoError(err)
	s.chConn = chConn
	_ = s.chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS transaction_events (
			transaction_id UInt64, transaction_no String, merchant_id UInt64, merchant_name String,
			card_number String, amount Int64, payment_method String, status String,
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (merchant_id, created_at)`)

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil, nil)
	s.merchantRepo = merchant_repo.NewRepositories(gormDB, nil)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	transactionRepos := repository.NewRepositories(gormDB, nil, nil, nil, nil, nil)
	transactionService := service.NewService(&service.Deps{
		Kafka: nil, Repositories: transactionRepos, MerchantAdapter: s.ts.MerchantAdapter,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log, Cache: cacheStore, AISecurityAdapter: nil,
	})

	// Seed Customer
	customer, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Txn", LastName: "Customer", Email: "txn.customer@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	cCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(customer.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = cCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: s.customerCardNumber, TotalBalance: 1000000})
	s.Require().NoError(err)

	// Seed Merchant
	owner, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merch", LastName: "Owner", Email: "merch.owner@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	merchant, err := s.merchantRepo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: int(owner.UserID), Name: "Txn Merchant",
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	s.merchantApiKey = merchant.ApiKey
	_, err = s.merchantRepo.UpdateMerchantStatus(context.Background(), &requests.UpdateMerchantStatusRequest{
		MerchantID: &s.merchantID, Status: "active",
	})
	s.Require().NoError(err)

	mCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(owner.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "321", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: mCard.CardNumber, TotalBalance: 0})
	s.Require().NoError(err)

	transactionHandlerGapi := handler.NewHandler(transactionService)
	chRepo := stats_repo.NewRepository(s.chConn)
	transactionStatsHandler := stats_handler.NewTransactionStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterTransactionCommandServiceServer(server, transactionHandlerGapi)
	pb.RegisterTransactionQueryServiceServer(server, transactionHandlerGapi)
	pbStats.RegisterTransactionStatsAmountServiceServer(server, transactionStatsHandler)
	pbStats.RegisterTransactionStatsMethodServiceServer(server, transactionStatsHandler)
	pbStats.RegisterTransactionStatsStatusServiceServer(server, transactionStatsHandler)
	pbAISecurity.RegisterAISecurityServiceServer(server, &mockAISecurityServer{})
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewTransactionCommandServiceClient(conn)
	s.queryClient = pb.NewTransactionQueryServiceClient(conn)
}

func (s *TransactionGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	s.redisClient.Close()
	if s.chConn != nil {
		s.chConn.Close()
	}
	s.ts.Teardown()
}

func (s *TransactionGapiTestSuite) Test1_CreateTransaction() {
	ctx := context.Background()
	res, err := s.commandClient.CreateTransaction(ctx, &pb.CreateTransactionRequest{
		CardNumber: s.customerCardNumber, Amount: 50000, PaymentMethod: "visa",
		MerchantId: int32(s.merchantID), TransactionTime: nil, ApiKey: s.merchantApiKey,
	})
	s.NoError(err)
	s.Equal("success", res.Status)
}

func (s *TransactionGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()
	res, err := s.queryClient.FindTransactionByMerchantId(ctx, &pb.FindTransactionByMerchantIdRequest{
		MerchantId: int32(s.merchantID),
	})
	s.NoError(err)
	s.Equal("success", res.Status)
}

func (s *TransactionGapiTestSuite) Test11_BulkOperations() {
	ctx := context.Background()
	_, err := s.commandClient.RestoreAllTransaction(ctx, &emptypb.Empty{})
	s.NoError(err)
	_, err = s.commandClient.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestTransactionGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionGapiTestSuite))
}
