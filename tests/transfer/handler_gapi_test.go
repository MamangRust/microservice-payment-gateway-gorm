package transfer_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TransferGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	redisClient   redis.UniversalClient
	chConn        clickhouse.Conn
	grpcServer    *grpc.Server
	commandClient pb.TransferCommandServiceClient
	queryClient   pb.TransferQueryServiceClient
	statsClient   pbStats.TransferStatsAmountServiceClient
	statusClient  pbStats.TransferStatsStatusServiceClient
	conn          *grpc.ClientConn
	repos         repository.Repositories
	userRepo      user_repo.UserCommandRepository
	cardRepo      *card_repo.Repositories
	saldoRepo     saldo_repo.Repositories

	senderCardNumber   string
	receiverCardNumber string
	transferID         int
}

func (s *TransferGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	// Run required migrations
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "transfer"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)

	chOpts, err := clickhouse.ParseDSN(s.ts.CHURL)
	s.Require().NoError(err)
	chConn, err := clickhouse.Open(chOpts)
	s.Require().NoError(err)
	s.chConn = chConn

	// Seed CH Schema
	err = s.chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS transfer_events (
			transfer_id UInt64,
			transfer_no String,
			source_card String,
			destination_card String,
			amount Int64,
			status String,
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (source_card, created_at);
	`)
	s.Require().NoError(err)

	// Repositories for seeding
	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = card_repo.NewRepositories(gormDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil)

	// Transfer repos
	s.repos = repository.NewRepositories(gormDB, s.saldoRepo, s.cardRepo.CardQuery)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	_, _ = observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	aiSecurityClient := pbAISecurity.NewAISecurityServiceClient(conn)

	transferService := service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     s.repos,
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: aiSecurityClient,
	})

	transferHandlerGapi := handler.NewHandler(transferService)

	// Stats Handler
	chRepo := stats_repo.NewRepository(s.chConn)
	transferStatsHandler := stats_handler.NewTransferStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterTransferCommandServiceServer(server, transferHandlerGapi)
	pb.RegisterTransferQueryServiceServer(server, transferHandlerGapi)
	pbStats.RegisterTransferStatsAmountServiceServer(server, transferStatsHandler)
	pbStats.RegisterTransferStatsStatusServiceServer(server, transferStatsHandler)
	pbAISecurity.RegisterAISecurityServiceServer(server, &mockAISecurityServer{})
	s.grpcServer = server

	go func() { _ = server.Serve(lis) }()

	s.commandClient = pb.NewTransferCommandServiceClient(conn)
	s.queryClient = pb.NewTransferQueryServiceClient(conn)
	s.statsClient = pbStats.NewTransferStatsAmountServiceClient(conn)
	s.statusClient = pbStats.NewTransferStatsStatusServiceClient(conn)

	// Seed Sender
	sender, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Sender",
		LastName:  "Gapi",
		Email:     "sender.gapi@test.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	sCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(sender.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "111",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCardNumber = sCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.senderCardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Seed Receiver
	receiver, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Receiver",
		LastName:  "Gapi",
		Email:     "receiver.gapi@test.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	rCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(receiver.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "222",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCardNumber = rCard.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.receiverCardNumber,
		TotalBalance: 0,
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.chConn != nil {
		s.chConn.Close()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *TransferGapiTestSuite) Test1_CreateTransfer() {
	ctx := context.Background()
	req := &pb.CreateTransferRequest{
		TransferFrom:   s.senderCardNumber,
		TransferTo:     s.receiverCardNumber,
		TransferAmount: 100000,
	}

	res, err := s.commandClient.CreateTransfer(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	s.transferID = int(res.Data.Id)

	// Verify balances
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.senderCardNumber)
	s.Equal(int64(900000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.receiverCardNumber)
	s.Equal(int64(100000), receiverSaldo.TotalBalance)
}

func (s *TransferGapiTestSuite) Test2_FindTransferById() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	res, err := s.queryClient.FindByIdTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	s.Equal(int32(s.transferID), res.Data.Id)
}

func (s *TransferGapiTestSuite) Test3_FindAllTransfers() {
	ctx := context.Background()
	res, err := s.queryClient.FindAllTransfer(ctx, &pb.FindAllTransferRequest{
		Page:     1,
		PageSize: 10,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
}

func (s *TransferGapiTestSuite) Test4_UpdateTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	req := &pb.UpdateTransferRequest{
		TransferId:     int32(s.transferID),
		TransferFrom:   s.senderCardNumber,
		TransferTo:     s.receiverCardNumber,
		TransferAmount: 150000, // Increase by 50000
	}

	res, err := s.commandClient.UpdateTransfer(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)

	// Verify adjusted balances (Sender 900k - 50k = 850k, Receiver 100k + 50k = 150k)
	senderSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.senderCardNumber)
	s.Equal(int64(850000), senderSaldo.TotalBalance)

	receiverSaldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.receiverCardNumber)
	s.Equal(int64(150000), receiverSaldo.TotalBalance)
}

func (s *TransferGapiTestSuite) Test5_TrashedTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.TrashedTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) Test6_RestoreTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.RestoreTransfer(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) Test7_PermanentDeleteTransfer() {
	s.Require().NotZero(s.transferID)
	ctx := context.Background()

	_, err := s.commandClient.DeleteTransferPermanent(ctx, &pb.FindByIdTransferRequest{
		TransferId: int32(s.transferID),
	})
	s.Require().NoError(err)
}

func (s *TransferGapiTestSuite) Test8_TransferStats_Amount() {
	ctx := context.Background()
	now := time.Now()

	err := s.chConn.Exec(ctx, "TRUNCATE TABLE transfer_events")
	s.Require().NoError(err)

	seedSQL := `INSERT INTO transfer_events (transfer_id, transfer_no, source_card, destination_card, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	// Sender data
	err = s.chConn.Exec(ctx, seedSQL, 1, "TR001", s.senderCardNumber, s.receiverCardNumber, 1000, "success", now)
	s.Require().NoError(err)
	// Receiver data
	err = s.chConn.Exec(ctx, seedSQL, 2, "TR002", "other_sender", s.senderCardNumber, 2000, "success", now)
	s.Require().NoError(err)
	// Failed data
	err = s.chConn.Exec(ctx, seedSQL, 3, "TR003", s.senderCardNumber, "some_receiver", 3000, "failed", now)
	s.Require().NoError(err)

	s.T().Run("MonthlyAmount", func(t *testing.T) {
		resp, err := s.statsClient.FindMonthlyTransferAmounts(ctx, &pb.FindYearTransferStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyAmount", func(t *testing.T) {
		resp, err := s.statsClient.FindYearlyTransferAmounts(ctx, &pb.FindYearTransferStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("MonthlyAmountBySender", func(t *testing.T) {
		resp, err := s.statsClient.FindMonthlyTransferAmountsBySenderCardNumber(ctx, &pb.FindByCardNumberTransferRequest{
			Year: int32(now.Year()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("MonthlyAmountByReceiver", func(t *testing.T) {
		resp, err := s.statsClient.FindMonthlyTransferAmountsByReceiverCardNumber(ctx, &pb.FindByCardNumberTransferRequest{
			Year: int32(now.Year()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyAmountBySender", func(t *testing.T) {
		resp, err := s.statsClient.FindYearlyTransferAmountsBySenderCardNumber(ctx, &pb.FindByCardNumberTransferRequest{
			Year: int32(now.Year()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyAmountByReceiver", func(t *testing.T) {
		resp, err := s.statsClient.FindYearlyTransferAmountsByReceiverCardNumber(ctx, &pb.FindByCardNumberTransferRequest{
			Year: int32(now.Year()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})
}

func (s *TransferGapiTestSuite) Test9_TransferStats_Status() {
	ctx := context.Background()
	now := time.Now()

	s.T().Run("MonthlySuccess", func(t *testing.T) {
		resp, err := s.statusClient.FindMonthlyTransferStatusSuccess(ctx, &pb.FindMonthlyTransferStatus{
			Year: int32(now.Year()), Month: int32(now.Month()),
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyFailed", func(t *testing.T) {
		resp, err := s.statusClient.FindYearlyTransferStatusFailed(ctx, &pb.FindYearTransferStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("MonthlySuccessByCard", func(t *testing.T) {
		resp, err := s.statusClient.FindMonthlyTransferStatusSuccessByCardNumber(ctx, &pb.FindMonthlyTransferStatusCardNumber{
			Year: int32(now.Year()), Month: int32(now.Month()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyFailedByCard", func(t *testing.T) {
		resp, err := s.statusClient.FindYearlyTransferStatusFailedByCardNumber(ctx, &pb.FindYearTransferStatusCardNumber{
			Year: int32(now.Year()), CardNumber: s.senderCardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})
}

func (s *TransferGapiTestSuite) Test11_BulkOperations() {
	ctx := context.Background()

	// Restore All
	_, err := s.commandClient.RestoreAllTransfer(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// Delete All Permanent
	_, err = s.commandClient.DeleteAllTransferPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestTransferGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferGapiTestSuite))
}

type mockAISecurityServer struct {
	pbAISecurity.UnimplementedAISecurityServiceServer
}

func (m *mockAISecurityServer) DetectFraud(ctx context.Context, in *pbAISecurity.FraudRequest) (*pbAISecurity.FraudResponse, error) {
	return &pbAISecurity.FraudResponse{
		TransactionId: in.TransactionId,
		RiskScore:     0.1,
		IsFraudulent:  false,
		Reason:        "Mocked safe transaction",
	}, nil
}
