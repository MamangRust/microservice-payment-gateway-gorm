package saldo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/saldo"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
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

type SaldoHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	chConn      clickhouse.Conn
	conn        *grpc.ClientConn
	router      *echo.Echo
	cardNumber  string
	userID      int
}

func (s *SaldoHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo"))

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
		CREATE TABLE IF NOT EXISTS saldo_events (
			card_number String, total_balance Int64, created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (card_number, created_at)`)

	userRepos := user_repo.NewRepositories(&user_repo.Deps{
		Db:              gormDB,
		RoleQueryClient: s.ts.RoleQueryClient,
		UserRoleClient:  s.ts.UserRoleClient,
	})
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	saldoService := service.NewService(&service.Deps{
		Repositories: saldoRepos, CardAdapter: s.ts.CardAdapter, Logger: log, Cache: cacheStore,
	})

	user, err := userRepos.UserCommand.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Saldo", LastName: "Handler", Email: "saldo.handler@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
	card, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: s.userID, CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	saldoHandler := handler.NewHandler(saldoService)
	chRepo := stats_repo.NewRepository(s.chConn)
	saldoStatsHandler := stats_handler.NewSaldoStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterSaldoCommandServiceServer(server, saldoHandler)
	pb.RegisterSaldoQueryServiceServer(server, saldoHandler)
	pbStats.RegisterSaldoStatsBalanceServiceServer(server, saldoStatsHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.router = echo.New()
	s.router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", fmt.Sprint(s.userID))
			return next(c)
		}
	})
	apiErrorHandler := app_errors.NewApiHandler(obs, log)
	api.RegisterSaldoHandler(&api.DepsSaldo{
		Client: conn, StatsClient: conn, E: s.router, Logger: log, Cache: cacheStore, ApiHandler: apiErrorHandler,
	})
}

func (s *SaldoHandlerApiTestSuite) TearDownSuite() {
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

func (s *SaldoHandlerApiTestSuite) Test1_CreateSaldo() {
	body, _ := json.Marshal(requests.CreateSaldoRequest{CardNumber: s.cardNumber, TotalBalance: 500000})
	request := httptest.NewRequest(http.MethodPost, "/api/saldo-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *SaldoHandlerApiTestSuite) Test2_FindById() {
	request := httptest.NewRequest(http.MethodGet, "/api/saldo-query/1", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *SaldoHandlerApiTestSuite) Test10_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/saldo-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/saldo-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestSaldoHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoHandlerApiTestSuite))
}
