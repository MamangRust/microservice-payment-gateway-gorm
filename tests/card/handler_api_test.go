package card_test

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
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	cardhandler "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
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

type CardApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	echoApp     *echo.Echo
	chConn      clickhouse.Conn
	userID      int
	cardID      int
}

func (s *CardApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card"))

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

	// Seed CH Schema
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
	userRepo := user_repo.NewRepositories(&user_repo.Deps{
		Db:              gormDB,
		RoleQueryClient: s.ts.RoleQueryClient,
		UserRoleClient:  s.ts.UserRoleClient,
	})

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

	cardGapiHandler := handler.NewHandler(cardService)

	// Stats Handler
	chRepo := stats_repo.NewRepository(s.chConn)
	cardStatsHandler := stats_handler.NewCardStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterCardQueryServiceServer(server, cardGapiHandler)
	pb.RegisterCardCommandServiceServer(server, cardGapiHandler)
	pb.RegisterCardDashboardServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsBalanceServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsTopupServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsTransactionServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsTransferServiceServer(server, cardStatsHandler)
	pbStats.RegisterCardStatsWithdrawServiceServer(server, cardStatsHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.echoApp = echo.New()
	obs, _ := observability.NewObservability("test", log)
	apiHandler := errors.NewApiHandler(obs, log)

	cardhandler.RegisterCardHandler(&cardhandler.DepsCard{
		Client:      conn,
		StatsClient: conn,
		E:           s.echoApp,
		Logger:      log,
		Cache:       cacheStore,
		ApiHandler:  apiHandler,
	})

	// Create user
	user, err := userRepo.UserCommand.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Api",
		LastName:  "Card",
		Email:     "api.card@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Auth Bypass
	s.echoApp.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("userID", strconv.Itoa(s.userID))
			return next(c)
		}
	})
}

func (s *CardApiTestSuite) TearDownSuite() {
	s.grpcServer.Stop()
	s.redisClient.Close()
	if s.chConn != nil {
		s.chConn.Close()
	}
	s.ts.Teardown()
}

func (s *CardApiTestSuite) Test1_CreateCard() {
	req := requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	}
	body, _ := json.Marshal(req)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/card-command/create", bytes.NewBuffer(body))
	httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	s.cardID = int(data["id"].(float64))
}

func (s *CardApiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.cardID)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card-query/%d", s.cardID), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	s.Equal(float64(s.cardID), data["id"].(float64))
}

func (s *CardApiTestSuite) Test3_UpdateCard() {
	s.Require().NotZero(s.cardID)
	req := requests.UpdateCardRequest{
		CardID:       s.cardID,
		UserID:       s.userID,
		CardType:     "credit",
		ExpireDate:   time.Now().AddDate(6, 0, 0),
		CVV:          "456",
		CardProvider: "MasterCard",
	}
	body, _ := json.Marshal(req)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/card-command/update/%d", s.cardID), bytes.NewBuffer(body))
	httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.cardID)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/card-command/trashed/%d", s.cardID), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/card-command/restore/%d", s.cardID), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.cardID)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/card-command/trashed/%d", s.cardID), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/card-command/permanent/%d", s.cardID), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test6_CardStats_MonthlyTopupAmount() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE topup_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO topup_events (topup_id, topup_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "TP001", "1234567890", 5000, "success", now)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card/stats/topup/monthly?year=%d", now.Year()), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test7_CardStats_Transaction() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE transaction_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO transaction_events (transaction_id, transaction_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "TX001", "1234567890", 1000, "success", now)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card/stats/transaction/monthly?year=%d", now.Year()), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test8_CardStats_Transfer() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE transfer_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO transfer_events (transfer_id, transfer_no, source_card, destination_card, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, 1, "TR001", "1234567890", "0987654321", 2000, "success", now)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card/stats/transfer/sender/monthly?year=%d", now.Year()), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test9_CardStats_Withdraw() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE withdraw_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`, 1, "WD001", "1234567890", 3000, "success", now)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card/stats/withdraw/monthly?year=%d", now.Year()), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test10_CardStats_Balance_Monthly() {
	ctx := context.Background()
	now := time.Now()
	_ = s.chConn.Exec(ctx, "TRUNCATE TABLE saldo_events")
	_ = s.chConn.Exec(ctx, `INSERT INTO saldo_events (card_number, total_balance, created_at) VALUES (?, ?, ?)`, "1234567890", 10000, now)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/card/stats/balance/monthly?year=%d", now.Year()), nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *CardApiTestSuite) Test14_BulkOperations() {
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/card-command/restore/all", nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/card-command/permanent/all", nil)
	s.echoApp.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestCardApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardApiTestSuite))
}
