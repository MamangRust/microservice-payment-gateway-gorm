package withdraw_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/withdraw"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WithdrawGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	redisClient   redis.UniversalClient
	chConn        clickhouse.Conn
	grpcServer    *grpc.Server
	commandClient pb.WithdrawCommandServiceClient
	queryClient   pb.WithdrawQueryServiceClient
	statsClient   pbStats.WithdrawStatsAmountServiceClient
	statusClient  pbStats.WithdrawStatsStatusServiceClient
	conn          *grpc.ClientConn
	repos         repository.Repositories
	userRepo      user_repo.UserCommandRepository
	cardRepo      card_repo.CardCommandRepository
	saldoRepo     saldo_repo.Repositories

	cardNumber string
	withdrawID int32
}

func (s *WithdrawGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	// Run required migrations
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "withdraw"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)

	chOpts, err := clickhouse.ParseDSN(s.ts.CHURL)
	s.Require().NoError(err)
	chConn, err := clickhouse.Open(chOpts)
	s.Require().NoError(err)
	s.chConn = chConn

	// Seed CH Schema
	err = s.chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS withdraw_events (
			withdraw_id UInt64,
			withdraw_no String,
			card_number String,
			card_type String,
			card_provider String,
			amount Int64,
			status String,
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (card_number, created_at);
	`)
	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	// Repositories for seeding and service dependencies
	userRepos := user_repo.NewUserCommandRepository(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil, nil)

	s.userRepo = userRepos
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	s.repos = repository.NewRepositories(gormDB, nil, nil, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	aiSecurityClient := pbAISecurity.NewAISecurityServiceClient(conn)

	// Seed User, Card, Saldo
	user, _ := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Withdraw", LastName: "Gapi", Email: "withdraw.gapi@test.com", Password: "password123",
	})
	card, _ := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "999", CardProvider: "visa",
	})
	s.cardNumber = card.CardNumber
	s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber: s.cardNumber, TotalBalance: 1000000,
	})

	withdrawService := service.NewService(&service.Deps{
		Kafka:             nil,
		Repositories:      s.repos,
		CardAdapter:       s.ts.CardAdapter,
		SaldoAdapter:      s.ts.SaldoAdapter,
		Logger:            log,
		Cache:             cacheStore,
		AISecurityAdapter: adapter.NewAISecurityAdapter(aiSecurityClient),
	})

	withdrawHandler := handler.NewHandler(withdrawService)

	// Stats Handler
	chRepo := stats_repo.NewRepository(s.chConn)
	withdrawStatsHandler := stats_handler.NewWithdrawStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterWithdrawCommandServiceServer(server, withdrawHandler)
	pb.RegisterWithdrawQueryServiceServer(server, withdrawHandler)
	pbStats.RegisterWithdrawStatsAmountServiceServer(server, withdrawStatsHandler)
	pbStats.RegisterWithdrawStatsStatusServiceServer(server, withdrawStatsHandler)
	pbAISecurity.RegisterAISecurityServiceServer(server, &mockAISecurityServer{})
	s.grpcServer = server

	go func() { _ = server.Serve(lis) }()

	s.commandClient = pb.NewWithdrawCommandServiceClient(conn)
	s.queryClient = pb.NewWithdrawQueryServiceClient(conn)
	s.statsClient = pbStats.NewWithdrawStatsAmountServiceClient(conn)
	s.statusClient = pbStats.NewWithdrawStatsStatusServiceClient(conn)
}

func (s *WithdrawGapiTestSuite) TearDownSuite() {
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

func (s *WithdrawGapiTestSuite) Test1_Create() {
	ctx := context.Background()

	createReq := &pb.CreateWithdrawRequest{
		CardNumber:     s.cardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   timestamppb.New(time.Now()),
	}
	res, err := s.commandClient.CreateWithdraw(ctx, createReq)
	s.NoError(err)
	s.Equal(int64(100000), res.Data.WithdrawAmount)

	s.withdrawID = res.Data.WithdrawId

	// Verify balance
	saldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.cardNumber)
	s.Equal(int64(900000), saldo.TotalBalance)
}

func (s *WithdrawGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	found, err := s.queryClient.FindByIdWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
	s.Equal(s.withdrawID, found.Data.WithdrawId)
}

func (s *WithdrawGapiTestSuite) Test3_Update() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	updateReq := &pb.UpdateWithdrawRequest{
		WithdrawId:     s.withdrawID,
		CardNumber:     s.cardNumber,
		WithdrawAmount: 150000,
		WithdrawTime:   timestamppb.New(time.Now()),
	}
	updated, err := s.commandClient.UpdateWithdraw(ctx, updateReq)
	s.NoError(err)
	s.Equal(int64(150000), updated.Data.WithdrawAmount)

	// Verify adjusted balance (900k - 50k = 850k)
	saldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.cardNumber)
	s.Equal(int64(850000), saldo.TotalBalance)
}

func (s *WithdrawGapiTestSuite) Test4_Trashed() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.TrashedWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func (s *WithdrawGapiTestSuite) Test5_Restore() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.RestoreWithdraw(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func (s *WithdrawGapiTestSuite) Test6_PermanentDelete() {
	s.Require().NotZero(s.withdrawID)

	ctx := context.Background()
	_, err := s.commandClient.DeleteWithdrawPermanent(ctx, &pb.FindByIdWithdrawRequest{WithdrawId: s.withdrawID})
	s.NoError(err)
}

func (s *WithdrawGapiTestSuite) Test7_WithdrawStats_Amount() {
	ctx := context.Background()
	now := time.Now()

	err := s.chConn.Exec(ctx, "TRUNCATE TABLE withdraw_events")
	s.Require().NoError(err)

	seedSQL := `INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, card_type, card_provider, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	// Success data
	err = s.chConn.Exec(ctx, seedSQL, 1, "WD001", s.cardNumber, "debit", "visa", 1000, "success", now)
	s.Require().NoError(err)
	// Failed data
	err = s.chConn.Exec(ctx, seedSQL, 2, "WD002", s.cardNumber, "debit", "visa", 2000, "failed", now)
	s.Require().NoError(err)

	s.T().Run("MonthlyAmount", func(t *testing.T) {
		resp, err := s.statsClient.FindMonthlyWithdraws(ctx, &pb.FindYearWithdrawStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyAmount", func(t *testing.T) {
		resp, err := s.statsClient.FindYearlyWithdraws(ctx, &pb.FindYearWithdrawStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("MonthlyAmountByCard", func(t *testing.T) {
		resp, err := s.statsClient.FindMonthlyWithdrawsByCardNumber(ctx, &pb.FindYearWithdrawCardNumber{
			Year: int32(now.Year()), CardNumber: s.cardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyAmountByCard", func(t *testing.T) {
		resp, err := s.statsClient.FindYearlyWithdrawsByCardNumber(ctx, &pb.FindYearWithdrawCardNumber{
			Year: int32(now.Year()), CardNumber: s.cardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})
}

func (s *WithdrawGapiTestSuite) Test8_WithdrawStats_Status() {
	ctx := context.Background()
	now := time.Now()

	s.T().Run("MonthlySuccess", func(t *testing.T) {
		resp, err := s.statusClient.FindMonthlyWithdrawStatusSuccess(ctx, &pb.FindMonthlyWithdrawStatus{
			Year: int32(now.Year()), Month: int32(now.Month()),
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyFailed", func(t *testing.T) {
		resp, err := s.statusClient.FindYearlyWithdrawStatusFailed(ctx, &pb.FindYearWithdrawStatus{Year: int32(now.Year())})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("MonthlySuccessByCard", func(t *testing.T) {
		resp, err := s.statusClient.FindMonthlyWithdrawStatusSuccessCardNumber(ctx, &pb.FindMonthlyWithdrawStatusCardNumber{
			Year: int32(now.Year()), Month: int32(now.Month()), CardNumber: s.cardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})

	s.T().Run("YearlyFailedByCard", func(t *testing.T) {
		resp, err := s.statusClient.FindYearlyWithdrawStatusFailedCardNumber(ctx, &pb.FindYearWithdrawStatusCardNumber{
			Year: int32(now.Year()), CardNumber: s.cardNumber,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Equal("success", resp.Status)
		s.NotEmpty(resp.Data)
	})
}

func (s *WithdrawGapiTestSuite) Test10_BulkOperations() {
	ctx := context.Background()

	// Restore All
	_, err := s.commandClient.RestoreAllWithdraw(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// Delete All Permanent
	_, err = s.commandClient.DeleteAllWithdrawPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
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

func TestWithdrawGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawGapiTestSuite))
}
