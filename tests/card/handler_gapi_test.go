package card_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CardGapiTestSuite struct {
	suite.Suite
	ts                *tests.TestSuite
	db                *gorm.DB
	redisClient       redis.UniversalClient
	grpcServer        *grpc.Server
	conn              *grpc.ClientConn
	chConn            clickhouse.Conn
	queryClient       pb.CardQueryServiceClient
	cmdClient         pb.CardCommandServiceClient
	statsClient       pbStats.CardStatsTopupServiceClient
	balanceClient     pbStats.CardStatsBalanceServiceClient
	transactionClient pbStats.CardStatsTransactionServiceClient
	transferClient    pbStats.CardStatsTransferServiceClient
	withdrawClient    pbStats.CardStatsWithdrawServiceClient
	userID            int
	cardID            int
}

func (s *CardGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations(
		"user", "role", "auth", "card", "merchant", "saldo", "transaction", "transfer", "withdraw", "topup",
	))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	chOpts, err := clickhouse.ParseDSN(s.ts.CHURL)
	s.Require().NoError(err)
	chConn, err := clickhouse.Open(chOpts)
	s.Require().NoError(err)
	s.chConn = chConn

	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS topup_events (topup_id UInt64, topup_no String, card_number String, card_type String, card_provider String, amount Int64, payment_method String, status String, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY (card_number, created_at)`,
		`CREATE TABLE IF NOT EXISTS transaction_events (transaction_id UInt64, transaction_no String, card_number String, amount Int64, status String, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY (card_number, created_at)`,
		`CREATE TABLE IF NOT EXISTS transfer_events (transfer_id UInt64, transfer_no String, source_card String, destination_card String, amount Int64, status String, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY (source_card, created_at)`,
		`CREATE TABLE IF NOT EXISTS saldo_events (card_number String, total_balance Int64, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY (card_number, created_at)`,
		`CREATE TABLE IF NOT EXISTS withdraw_events (withdraw_id UInt64, withdraw_no String, card_number String, amount Int64, status String, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY (card_number, created_at)`,
	} {
		err = s.chConn.Exec(context.Background(), ddl)
		s.Require().NoError(err)
	}

	repos := repository.NewRepositories(gormDB, nil)
	userRepo := user_repo.NewRepositories(&user_repo.Deps{Db: gormDB, RoleQueryClient: s.ts.RoleQueryClient, UserRoleClient: s.ts.UserRoleClient})

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	cardService := service.NewService(&service.Deps{
		Repositories: repos,
		UserAdapter:  s.ts.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	cardHandler := handler.NewHandler(cardService)
	chRepo := stats_repo.NewRepository(s.chConn)
	cardStatsHandler := stats_handler.NewCardStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterCardQueryServiceServer(server, cardHandler)
	pb.RegisterCardCommandServiceServer(server, cardHandler)
	pbStats.RegisterCardStatsTopupServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsBalanceServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsTransactionServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsTransferServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsWithdrawServiceServer(server, cardStatsHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.queryClient = pb.NewCardQueryServiceClient(conn)
	s.cmdClient = pb.NewCardCommandServiceClient(conn)
	s.statsClient = pbStats.NewCardStatsTopupServiceClient(conn)
	s.balanceClient = pbStats.NewCardStatsBalanceServiceClient(conn)
	s.transactionClient = pbStats.NewCardStatsTransactionServiceClient(conn)
	s.transferClient = pbStats.NewCardStatsTransferServiceClient(conn)
	s.withdrawClient = pbStats.NewCardStatsWithdrawServiceClient(conn)

	user, err := userRepo.UserCommand.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Gapi",
		LastName:  "Card",
		Email:     "gapi.card@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *CardGapiTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	if s.chConn != nil {
		s.chConn.Close()
	}
	s.ts.Teardown()
}

func (s *CardGapiTestSuite) Test1_CreateCard() {
	res, err := s.cmdClient.CreateCard(context.Background(), &pb.CreateCardRequest{
		UserId:       int32(s.userID),
		CardType:     "debit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(5, 0, 0)),
		Cvv:          "123",
		CardProvider: "Visa",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.cardID = int(res.Data.Id)
}

func (s *CardGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.cardID)
	res, err := s.queryClient.FindByIdCard(context.Background(), &pb.FindByIdCardRequest{CardId: int32(s.cardID)})
	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(s.cardID), res.Data.Id)
}

func (s *CardGapiTestSuite) Test3_UpdateCard() {
	s.Require().NotZero(s.cardID)
	res, err := s.cmdClient.UpdateCard(context.Background(), &pb.UpdateCardRequest{
		CardId:       int32(s.cardID),
		UserId:       int32(s.userID),
		CardType:     "credit",
		ExpireDate:   timestamppb.New(time.Now().AddDate(6, 0, 0)),
		Cvv:          "456",
		CardProvider: "MasterCard",
	})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
}

func (s *CardGapiTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()
	trashRes, err := s.cmdClient.TrashedCard(ctx, &pb.FindByIdCardRequest{CardId: int32(s.cardID)})
	s.NoError(err)
	s.Equal("success", trashRes.Status)
	restoreRes, err := s.cmdClient.RestoreCard(ctx, &pb.FindByIdCardRequest{CardId: int32(s.cardID)})
	s.NoError(err)
	s.Equal("success", restoreRes.Status)
}

func (s *CardGapiTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()
	_, _ = s.cmdClient.TrashedCard(ctx, &pb.FindByIdCardRequest{CardId: int32(s.cardID)})
	delRes, err := s.cmdClient.DeleteCardPermanent(ctx, &pb.FindByIdCardRequest{CardId: int32(s.cardID)})
	s.NoError(err)
	s.Equal("success", delRes.Status)
}

func (s *CardGapiTestSuite) Test6_CardStats_MonthlyTopupAmount() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE topup_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO topup_events (topup_id, topup_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "TP001", "1234567890", 5000, "success", now)
	resp, err := s.statsClient.FindMonthlyTopupAmount(ctx, &pb.FindYearAmount{Year: int32(now.Year())})
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func (s *CardGapiTestSuite) Test7_CardStats_Transaction() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE transaction_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO transaction_events (transaction_id, transaction_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "TX001", "1234567890", 1000, "success", now)
	resp, err := s.transactionClient.FindMonthlyTransactionAmount(ctx, &pb.FindYearAmount{Year: int32(now.Year())})
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func (s *CardGapiTestSuite) Test8_CardStats_Transfer() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE transfer_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO transfer_events (transfer_id, transfer_no, source_card, destination_card, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, 1, "TR001", "1234567890", "0987654321", 2000, "success", now)
	resp, err := s.transferClient.FindMonthlyTransferSenderAmount(ctx, &pb.FindYearAmount{Year: int32(now.Year())})
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func (s *CardGapiTestSuite) Test9_CardStats_Withdraw() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE withdraw_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "WD001", "1234567890", 3000, "success", now)
	resp, err := s.withdrawClient.FindMonthlyWithdrawAmount(ctx, &pb.FindYearAmount{Year: int32(now.Year())})
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func (s *CardGapiTestSuite) Test11_CardStats_Balance_Full() {
	ctx := context.Background()
	now := time.Now()
	cardNumber := "1234567890"
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE saldo_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO saldo_events (card_number, total_balance, created_at) VALUES (?, ?, ?)`, cardNumber, 10000, now)
	respM, err := s.balanceClient.FindMonthlyBalance(ctx, &pbStats.FindYearBalance{Year: int32(now.Year())})
	s.Require().NoError(err)
	s.Equal("success", respM.Status)
	respY, err := s.balanceClient.FindYearlyBalance(ctx, &pbStats.FindYearBalance{Year: int32(now.Year())})
	s.Require().NoError(err)
	s.Equal("success", respY.Status)
}

func (s *CardGapiTestSuite) Test12_CardStats_Topup_Full() {
	now := time.Now()
	cardNumber := "1234567890"
	_, err := s.statsClient.FindYearlyTopupAmount(context.Background(), &pb.FindYearAmount{Year: int32(now.Year())})
	s.NoError(err)
	_, err = s.statsClient.FindMonthlyTopupAmountByCardNumber(context.Background(), &pb.FindYearAmountCardNumber{Year: int32(now.Year()), CardNumber: cardNumber})
	s.NoError(err)
	_, err = s.statsClient.FindYearlyTopupAmountByCardNumber(context.Background(), &pb.FindYearAmountCardNumber{Year: int32(now.Year()), CardNumber: cardNumber})
	s.NoError(err)
}

func (s *CardGapiTestSuite) Test16_BulkOperations() {
	ctx := context.Background()
	_, err := s.cmdClient.RestoreAllCard(ctx, &emptypb.Empty{})
	s.NoError(err)
	_, err = s.cmdClient.DeleteAllCardPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestCardGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardGapiTestSuite))
}
