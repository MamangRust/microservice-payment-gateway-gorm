package topup_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/topup"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/handler"
	topup_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/service"
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

type TopupGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	db            *gorm.DB
	redisClient   redis.UniversalClient
	chConn        clickhouse.Conn
	grpcServer    *grpc.Server
	commandClient pb.TopupCommandServiceClient
	queryClient   pb.TopupQueryServiceClient
	statsClient   pbStats.TopupStatsAmountServiceClient
	methodClient  pbStats.TopupStatsMethodServiceClient
	statusClient  pbStats.TopupStatsStatusServiceClient
	conn          *grpc.ClientConn
	userRepo      user_repo.UserCommandRepository
	cardRepo      card_repo.CardCommandRepository
	saldoRepo     saldo_repo.Repositories
	topupRepo     topup_repo.Repositories
	cardNumber    string
	topupID       int32
}

func (s *TopupGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "topup"))

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
	_ = s.chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS topup_events (
			topup_id UInt64, topup_no String, card_number String, card_type String,
			card_provider String, amount Int64, payment_method String, status String,
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (card_number, created_at)`)

	userRepos := user_repo.NewRepositories(&user_repo.Deps{Db: gormDB, RoleQueryClient: s.ts.RoleQueryClient, UserRoleClient: s.ts.UserRoleClient})
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil, nil)

	s.topupRepo = topup_repo.NewRepositories(gormDB, nil, nil, nil, nil)
	s.userRepo = userRepos.UserCommand
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	topupService := service.NewService(&service.Deps{
		Kafka: nil, Cache: cacheStore, Repositories: s.topupRepo,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log,
	})

	user, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Topup", LastName: "Gapi", Email: "topup.gapi@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "444", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber: s.cardNumber, TotalBalance: 0,
	})
	s.Require().NoError(err)

	topupHandler := gapi.NewHandler(topupService)
	chRepo := stats_repo.NewRepository(s.chConn)
	topupStatsHandler := stats_handler.NewTopupStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterTopupCommandServiceServer(server, topupHandler)
	pb.RegisterTopupQueryServiceServer(server, topupHandler)
	pbStats.RegisterTopupStatsAmountServiceServer(server, topupStatsHandler)
	pbStats.RegisterTopupStatsMethodServiceServer(server, topupStatsHandler)
	pbStats.RegisterTopupStatsStatusServiceServer(server, topupStatsHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewTopupCommandServiceClient(conn)
	s.queryClient = pb.NewTopupQueryServiceClient(conn)
	s.statsClient = pbStats.NewTopupStatsAmountServiceClient(conn)
	s.methodClient = pbStats.NewTopupStatsMethodServiceClient(conn)
	s.statusClient = pbStats.NewTopupStatsStatusServiceClient(conn)
}

func (s *TopupGapiTestSuite) TearDownSuite() {
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

func (s *TopupGapiTestSuite) Test1_Create() {
	ctx := context.Background()
	res, err := s.commandClient.CreateTopup(ctx, &pb.CreateTopupRequest{
		CardNumber: s.cardNumber, TopupAmount: 100000, TopupMethod: "bri",
	})
	s.NoError(err)
	s.Equal(int64(100000), res.Data.TopupAmount)
	s.topupID = res.Data.Id
	saldo, _ := s.saldoRepo.FindByCardNumber(ctx, s.cardNumber)
	s.Equal(int64(100000), saldo.TotalBalance)
}

func (s *TopupGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.topupID)
	found, err := s.queryClient.FindByIdTopup(context.Background(), &pb.FindByIdTopupRequest{TopupId: s.topupID})
	s.NoError(err)
	s.Equal(s.topupID, found.Data.Id)
}

func (s *TopupGapiTestSuite) Test3_Update() {
	s.Require().NotZero(s.topupID)
	ctx := context.Background()
	updated, err := s.commandClient.UpdateTopup(ctx, &pb.UpdateTopupRequest{
		TopupId: s.topupID, CardNumber: s.cardNumber, TopupAmount: 150000, TopupMethod: "bri",
	})
	s.NoError(err)
	s.Equal(int64(150000), updated.Data.TopupAmount)
}

func (s *TopupGapiTestSuite) Test4_Trashed() {
	s.Require().NotZero(s.topupID)
	_, err := s.commandClient.TrashedTopup(context.Background(), &pb.FindByIdTopupRequest{TopupId: s.topupID})
	s.NoError(err)
}

func (s *TopupGapiTestSuite) Test5_Restore() {
	s.Require().NotZero(s.topupID)
	_, err := s.commandClient.RestoreTopup(context.Background(), &pb.FindByIdTopupRequest{TopupId: s.topupID})
	s.NoError(err)
}

func (s *TopupGapiTestSuite) Test6_DeletePermanent() {
	s.Require().NotZero(s.topupID)
	_, err := s.commandClient.DeleteTopupPermanent(context.Background(), &pb.FindByIdTopupRequest{TopupId: s.topupID})
	s.NoError(err)
}

func (s *TopupGapiTestSuite) Test11_BulkOperations() {
	ctx := context.Background()
	_, err := s.commandClient.RestoreAllTopup(ctx, &emptypb.Empty{})
	s.NoError(err)
	_, err = s.commandClient.DeleteAllTopupPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestTopupGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupGapiTestSuite))
}
