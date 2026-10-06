package transfer_test

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
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/transfer"
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

type TransferHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	chConn      clickhouse.Conn
	conn        *grpc.ClientConn
	router      *echo.Echo
	userRepo    user_repo.UserCommandRepository
	cardRepo    card_repo.Repositories
	saldoRepo   saldo_repo.Repositories
	senderCard  string
	receiverCard string
	userID      int
}

func (s *TransferHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "transfer"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	chOpts, err := clickhouse.ParseDSN(s.ts.CHURL)
	s.Require().NoError(err)
	chConn, err := clickhouse.Open(chOpts)
	s.Require().NoError(err)
	s.chConn = chConn
	_ = s.chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS transfer_events (
			transfer_id UInt64, transfer_no String, source_card String, destination_card String,
			amount Int64, status String, created_at DateTime DEFAULT now()
		) ENGINE = MergeTree() ORDER BY (source_card, created_at)`)

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, s.ts.UserClient)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, s.ts.CardClient, s.ts.CardClient)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	transferRepos := repository.NewRepositories(gormDB, s.ts.SaldoClient, s.ts.SaldoClient, s.ts.CardClient, s.ts.CardClient)
	transferService := service.NewService(&service.Deps{
		Kafka: nil, Repositories: transferRepos, SaldoAdapter: s.ts.SaldoAdapter,
		CardAdapter: s.ts.CardAdapter, Logger: log, Cache: cacheStore,
	})

	// Seed sender + receiver
	sender, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Sender", LastName: "Handler", Email: "sender.handler@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(sender.UserID)
	sCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: s.userID, CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "111", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCard = sCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: s.senderCard, TotalBalance: 1000000})
	s.Require().NoError(err)

	receiver, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Receiver", LastName: "Handler", Email: "receiver.handler@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	rCard, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID: int(receiver.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "222", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCard = rCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{CardNumber: s.receiverCard, TotalBalance: 0})
	s.Require().NoError(err)

	transferHandler := handler.NewHandler(transferService)
	chRepo := stats_repo.NewRepository(s.chConn)
	transferStatsHandler := stats_handler.NewTransferStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterTransferCommandServiceServer(server, transferHandler)
	pb.RegisterTransferQueryServiceServer(server, transferHandler)
	pbStats.RegisterTransferStatsAmountServiceServer(server, transferStatsHandler)
	pbStats.RegisterTransferStatsStatusServiceServer(server, transferStatsHandler)
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
	api.RegisterTransferHandler(&api.DepsTransfer{
		Client: conn, StatsClient: conn, E: s.router, Logger: log, Cache: cacheStore, ApiHandler: apiErrorHandler,
	})
}

func (s *TransferHandlerApiTestSuite) TearDownSuite() {
	if s.conn != nil { s.conn.Close() }
	if s.grpcServer != nil { s.grpcServer.Stop() }
	s.redisClient.Close()
	if s.chConn != nil { s.chConn.Close() }
	s.ts.Teardown()
}

func (s *TransferHandlerApiTestSuite) Test1_CreateTransfer() {
	body, _ := json.Marshal(requests.CreateTransferRequest{
		TransferFrom: s.senderCard, TransferTo: s.receiverCard, TransferAmount: 50000,
	})
	request := httptest.NewRequest(http.MethodPost, "/api/transfer-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "transfer-create-handler-1")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *TransferHandlerApiTestSuite) Test10_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/transfer-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/transfer-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestTransferHandlerApiSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransferHandlerApiTestSuite))
}
