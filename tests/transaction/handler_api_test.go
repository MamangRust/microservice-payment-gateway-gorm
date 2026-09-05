package transaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pb_merchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api_transaction "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/transaction"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis"
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
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransactionHandlerTestSuite struct {
	suite.Suite
	ts             *tests.TestSuite
	db             *gorm.DB
	redisClient    redis.UniversalClient
	grpcServer     *grpc.Server
	chConn         clickhouse.Conn
	commandClient  pb.TransactionCommandServiceClient
	queryClient    pb.TransactionQueryServiceClient
	merchantClient pb_merchant.MerchantCommandServiceClient
	conn           *grpc.ClientConn
	router         *echo.Echo
	userRepo       user_repo.UserCommandRepository
	cardRepo       card_repo.Repositories
	saldoRepo      saldo_repo.Repositories
	merchantRepo   merchant_repo.Repositories
	customerCardNumber string
	merchantApiKey string
	merchantID     int
	merchantCardNumber string
	transactionID  int
}

func (s *TransactionHandlerTestSuite) SetupSuite() {
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
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil)
	s.merchantRepo = merchant_repo.NewRepositories(gormDB, nil)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	cardRepoWrapper := &transactionCardRepo{
		query: s.cardRepo.CardQuery, command: s.cardRepo.CardCommand,
	}
	transactionRepos := repository.NewRepositories(gormDB, s.saldoRepo, cardRepoWrapper, s.merchantRepo)
	transactionService := service.NewService(&service.Deps{
		Kafka: nil, Repositories: transactionRepos, MerchantAdapter: s.ts.MerchantAdapter,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log, Cache: cacheStore, AISecurityClient: nil,
	})

	// Seed Customer
	customer, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Transaction", LastName: "Customer", Email: "customer@transaction.com", Password: "password123",
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
		FirstName: "Merchant", LastName: "Owner", Email: "merchant.owner@transaction.com", Password: "password123",
	})
	s.Require().NoError(err)
	merchant, err := s.merchantRepo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: int(owner.UserID), Name: "Transaction Merchant",
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	_, err = s.merchantRepo.UpdateMerchantStatus(context.Background(), &requests.UpdateMerchantStatusRequest{
		MerchantID: &s.merchantID, Status: "active",
	})
	s.Require().NoError(err)
	mFull, _ := s.merchantRepo.FindByMerchantId(context.Background(), s.merchantID)
	s.merchantApiKey = mFull.ApiKey

	mCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(owner.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "321", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.merchantCardNumber = mCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: s.merchantCardNumber, TotalBalance: 0})
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
	s.merchantClient = pb_merchant.NewMerchantCommandServiceClient(conn)

	s.router = echo.New()
	apiErrorHandler := errors.NewApiHandler(obs, log)
	merchantCache := mencache.NewMerchantCache(cacheStore)
	merchantCache.SetMerchantCache(context.Background(), strconv.Itoa(s.merchantID), s.merchantApiKey)

	api_transaction.RegisterTransactionHandler(&api_transaction.DepsTransaction{
		Client: conn, StatsClient: conn, E: s.router, Kafka: nil, Logger: log, Cache: cacheStore,
		ApiHandler: apiErrorHandler, CacheApiGateway: mencache.NewCacheApiGateway(cacheStore), AISecurity: conn,
	})
}

func (s *TransactionHandlerTestSuite) TearDownSuite() {
	if s.conn != nil { s.conn.Close() }
	if s.grpcServer != nil { s.grpcServer.Stop() }
	s.redisClient.Close()
	if s.chConn != nil { s.chConn.Close() }
	s.ts.Teardown()
}

func (s *TransactionHandlerTestSuite) Test1_CreateTransaction() {
	createReq := map[string]interface{}{
		"card_number": s.customerCardNumber, "amount": 50000, "payment_method": "visa",
		"merchant_id": s.merchantID, "transaction_time": time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(createReq)
	request := httptest.NewRequest(http.MethodPost, "/api/transaction-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set("X-API-Key", s.merchantApiKey)
	request.Header.Set("Idempotency-Key", "transaction-create-handler-1")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.transactionID = int(data["id"].(float64))

	customerSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.customerCardNumber)
	s.Equal(int64(950000), customerSaldo.TotalBalance)
	merchantSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.merchantCardNumber)
	s.Equal(int64(50000), merchantSaldo.TotalBalance)
}

func (s *TransactionHandlerTestSuite) Test2_FindTransactionById() {
	s.Require().NotZero(s.transactionID)
	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/transaction-query/%d", s.transactionID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *TransactionHandlerTestSuite) Test3_FindAllTransactions() {
	request := httptest.NewRequest(http.MethodGet, "/api/transaction-query?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *TransactionHandlerTestSuite) Test11_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/transaction-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/transaction-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestTransactionHandlerSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransactionHandlerTestSuite))
}
