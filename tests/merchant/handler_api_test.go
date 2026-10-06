package merchant_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
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

type MerchantHandlerTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	chConn      clickhouse.Conn
	conn        *grpc.ClientConn
	router      *echo.Echo
	userRepo    user_repo.UserCommandRepository
	userID      int
	merchantID  int
}

func (s *MerchantHandlerTestSuite) SetupSuite() {
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
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	merchantService := service.NewService(&service.Deps{
		Kafka: nil, Repositories: repos, UserAdapter: s.ts.UserAdapter, Logger: log, Cache: cacheStore,
	})

	user, err := s.userRepo.CreateUser(s.ts.Ctx, &requests.CreateUserRequest{
		FirstName: "Handler", LastName: "Merchant", Email: "handler.merchant@example.com", Password: "password123",
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

	s.router = echo.New()
	apiErrorHandler := app_errors.NewApiHandler(obs, log)
	api.RegisterMerchantHandler(&api.DepsMerchant{
		Client: conn, StatsClient: conn, E: s.router, Logger: log, Cache: cacheStore, ApiHandler: apiErrorHandler,
	})
}

func (s *MerchantHandlerTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	if s.chConn != nil { s.chConn.Close() }
	s.ts.Teardown()
}

func (s *MerchantHandlerTestSuite) Test1_CreateMerchant() {
	body, _ := json.Marshal(requests.CreateMerchantRequest{Name: "Handler Merchant", UserID: s.userID})
	request := httptest.NewRequest(http.MethodPost, "/api/merchant-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var createRes map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createRes)
	data := createRes["data"].(map[string]interface{})
	s.merchantID = int(data["id"].(float64))
}

func (s *MerchantHandlerTestSuite) Test2_FindMerchantById() {
	s.Require().NotZero(s.merchantID)
	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/merchant-query/%d", s.merchantID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *MerchantHandlerTestSuite) Test10_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/merchant-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/merchant-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestMerchantHandlerSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(MerchantHandlerTestSuite))
}
