package topup_test

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
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/topup"
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

type TopupHandlerTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	chConn      clickhouse.Conn
	conn        *grpc.ClientConn
	router      *echo.Echo
	userRepo    user_repo.UserCommandRepository
	cardRepo    card_repo.CardCommandRepository
	saldoRepo   saldo_repo.Repositories
	topupRepo   topup_repo.Repositories
	cardNumber  string
	topupID     int
	userID      int
}

func (s *TopupHandlerTestSuite) SetupSuite() {
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

	userRepos := user_repo.NewRepositories(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil)

	cardAdapter := &topupCardRepoAdapter{
		CardQueryRepository: cardRepos.CardQuery, CardCommandRepository: cardRepos.CardCommand,
	}
	s.topupRepo = topup_repo.NewRepositories(gormDB, cardAdapter, saldoRepos)
	s.userRepo = userRepos.UserCommand()
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	topupService := service.NewService(&service.Deps{
		Kafka: nil, Cache: cacheStore, Repositories: s.topupRepo,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log,
	})

	user, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Topup", LastName: "Owner", Email: "topup.handler@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	card, err := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: s.userID, CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber: s.cardNumber, TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	topupGapiHandler := gapi.NewHandler(topupService)
	chRepo := stats_repo.NewRepository(s.chConn)
	topupStatsHandler := stats_handler.NewTopupStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterTopupCommandServiceServer(server, topupGapiHandler)
	pb.RegisterTopupQueryServiceServer(server, topupGapiHandler)
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

	s.router = echo.New()
	s.router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", fmt.Sprint(s.userID))
			return next(c)
		}
	})
	apiErrorHandler := app_errors.NewApiHandler(obs, log)
	api.RegisterTopupHandler(&api.DepsTopup{
		Client: conn, StatsClient: conn, E: s.router, Logger: log, Cache: cacheStore, ApiHandler: apiErrorHandler,
	})
}

func (s *TopupHandlerTestSuite) TearDownSuite() {
	if s.conn != nil { s.conn.Close() }
	if s.grpcServer != nil { s.grpcServer.Stop() }
	s.redisClient.Close()
	if s.chConn != nil { s.chConn.Close() }
	s.ts.Teardown()
}

func (s *TopupHandlerTestSuite) Test1_CreateTopup() {
	body, _ := json.Marshal(requests.CreateTopupRequest{
		CardNumber: s.cardNumber, TopupAmount: 100000, TopupMethod: "visa",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/topup-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "topup-create-handler-1")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
	var createRes map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createRes)
	data := createRes["data"].(map[string]interface{})
	s.topupID = int(data["id"].(float64))
}

func (s *TopupHandlerTestSuite) Test2_FindById() {
	s.Require().NotZero(s.topupID)
	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/topup-query/%d", s.topupID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *TopupHandlerTestSuite) Test10_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/topup-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/topup-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestTopupHandlerSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TopupHandlerTestSuite))
}
