package merchant_test

import (
	"context"
	"net"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
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
)

type MerchantGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	db            *gorm.DB
	redisClient   redis.UniversalClient
	grpcServer    *grpc.Server
	chConn        clickhouse.Conn
	commandClient pb.MerchantCommandServiceClient
	queryClient   pb.MerchantQueryServiceClient
	conn          *grpc.ClientConn
	userRepo      user_repo.UserCommandRepository
	userID        int
	merchantID    int
}

func (s *MerchantGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "transaction"))
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
		CREATE TABLE IF NOT EXISTS transaction_events (
			transaction_id UInt64, transaction_no String, merchant_id UInt64, merchant_name String,
			apikey String, apikey_name String, amount Int64, payment_method String, status String,
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (merchant_id, created_at)`)
	repos := repository.NewRepositories(gormDB, nil)
	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	merchantService := service.NewService(&service.Deps{
		Kafka: nil, Repositories: repos, UserAdapter: s.ts.UserAdapter, Logger: log, Cache: cacheStore,
	})
	user, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Gapi", LastName: "Merchant", Email: "gapi.merchant@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
	merchantHandler := handler.NewHandler(merchantService)
	chRepo := stats_repo.NewRepository(s.chConn)
	merchantStatsHandler := stats_handler.NewMerchantStatsHandler(chRepo, log)
	server := grpc.NewServer()
	pb.RegisterMerchantCommandServiceServer(server, merchantHandler)
	pb.RegisterMerchantQueryServiceServer(server, merchantHandler)
	pbStats.RegisterMerchantStatsAmountServiceServer(server, merchantStatsHandler)
	pbStats.RegisterMerchantStatsMethodServiceServer(server, merchantStatsHandler)
	pbStats.RegisterMerchantStatsTotalAmountServiceServer(server, merchantStatsHandler)
	pb.RegisterMerchantTransactionServiceServer(server, merchantStatsHandler)
	s.grpcServer = server
	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewMerchantCommandServiceClient(conn)
	s.queryClient = pb.NewMerchantQueryServiceClient(conn)
}

func (s *MerchantGapiTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	if s.chConn != nil {
		s.chConn.Close()
	}
	s.ts.Teardown()
}

func (s *MerchantGapiTestSuite) Test1_CreateMerchant() {
	res, err := s.commandClient.CreateMerchant(context.Background(), &pb.CreateMerchantRequest{
		Name: "Gapi Merchant", UserId: int32(s.userID),
	})
	s.NoError(err)
	s.Equal("Gapi Merchant", res.Data.Name)
	s.merchantID = int(res.Data.Id)
}

func (s *MerchantGapiTestSuite) Test2_FindMerchantById() {
	s.Require().NotZero(s.merchantID)
	found, err := s.queryClient.FindByIdMerchant(context.Background(), &pb.FindByIdMerchantRequest{MerchantId: int32(s.merchantID)})
	s.NoError(err)
	s.Equal(int32(s.merchantID), found.Data.Id)
}

func (s *MerchantGapiTestSuite) Test10_BulkOperations() {
	ctx := context.Background()
	_, err := s.commandClient.RestoreAllMerchant(ctx, &emptypb.Empty{})
	s.NoError(err)
	_, err = s.commandClient.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestMerchantGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGapiTestSuite))
}
