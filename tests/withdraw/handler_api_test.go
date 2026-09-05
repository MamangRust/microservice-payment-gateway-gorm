package withdraw_test

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
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	pbStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	withdrawhandler "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/withdraw"
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
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WithdrawHandlerTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	redisClient   redis.UniversalClient
	chConn        clickhouse.Conn
	grpcServer    *grpc.Server
	commandClient pb.WithdrawCommandServiceClient
	queryClient   pb.WithdrawQueryServiceClient
	conn          *grpc.ClientConn
	router        *echo.Echo
	repos         repository.Repositories
	userRepo      user_repo.UserCommandRepository
	cardRepo      card_repo.CardCommandRepository
	saldoRepo     saldo_repo.Repositories

	customerCardNumber string
	customerUserID     int
	withdrawID         int
}

func (s *WithdrawHandlerTestSuite) SetupSuite() {
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

	// Repositories for seeding and service dependencies
	userRepos := user_repo.NewUserCommandRepository(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil)

	s.userRepo = userRepos
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	s.repos = repository.NewRepositories(gormDB, cardRepos.CardQuery, saldoRepos)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	withdrawService := service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     s.repos,
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: nil,
	})

	// Seed Customer
	customer, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Withdraw",
		LastName:  "Customer",
		Email:     "withdraw@test.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.customerUserID = int(customer.UserID)

	card, err := s.cardRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.customerUserID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "999",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = card.CardNumber

	_, err = s.saldoRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.customerCardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	withdrawHandler := handler.NewHandler(withdrawService)

	// Stats Handler
	chRepo := stats_repo.NewRepository(s.chConn)
	withdrawStatsHandler := stats_handler.NewWithdrawStatsHandler(chRepo, log)

	server := grpc.NewServer()
	pb.RegisterWithdrawCommandServiceServer(server, withdrawHandler)
	pb.RegisterWithdrawQueryServiceServer(server, withdrawHandler)
	pbStats.RegisterWithdrawStatsAmountServiceServer(server, withdrawStatsHandler)
	pbStats.RegisterWithdrawStatsStatusServiceServer(server, withdrawStatsHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	go func() {
		_ = server.Serve(lis)
	}()

	// Create gRPC Client
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewWithdrawCommandServiceClient(conn)
	s.queryClient = pb.NewWithdrawQueryServiceClient(conn)

	// Setup Echo
	s.router = echo.New()
	// Mimic the production JWT auth middleware (WebSecurityConfig): command
	// handlers read c.Get("user_id") and reject with 401 when it is absent.
	// The seeded card belongs to s.customerUserID, so assert ownership with it.
	s.router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", fmt.Sprint(s.customerUserID))
			return next(c)
		}
	})
	apiErrorHandler := app_errors.NewApiHandler(obs, log)

	withdrawhandler.RegisterWithdrawHandler(&withdrawhandler.DepsWithdraw{
		Client:      conn,
		StatsClient: conn, // Using same conn for stats
		E:           s.router,
		Logger:      log,
		Cache:       cacheStore,
		ApiHandler:  apiErrorHandler,
	})
}

func (s *WithdrawHandlerTestSuite) TearDownSuite() {
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

func (s *WithdrawHandlerTestSuite) Test1_CreateWithdraw() {
	req := &requests.CreateWithdrawRequest{
		CardNumber:     s.customerCardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest(http.MethodPost, "/api/withdraw-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", "withdraw-create-handler-1")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	var createRes map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createRes)
	withdrawData := createRes["data"].(map[string]interface{})
	s.withdrawID = int(withdrawData["id"].(float64))

	// Verify balance
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.customerCardNumber)
	s.Equal(int64(900000), customerSaldo.TotalBalance)
}

func (s *WithdrawHandlerTestSuite) Test2_FindWithdrawById() {
	s.Require().NotZero(s.withdrawID)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw-query/%d", s.withdrawID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *WithdrawHandlerTestSuite) Test3_FindAllWithdraws() {
	request := httptest.NewRequest(http.MethodGet, "/api/withdraw-query?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *WithdrawHandlerTestSuite) Test4_UpdateWithdraw() {
	s.Require().NotZero(s.withdrawID)

	req := &requests.UpdateWithdrawRequest{
		WithdrawID:     &s.withdrawID,
		CardNumber:     s.customerCardNumber,
		WithdrawAmount: 150000, // Increase by 50000
		WithdrawTime:   time.Now(),
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/withdraw-command/update/%d", s.withdrawID), bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	// Verify adjusted balance (900k - 50k = 850k)
	customerSaldo, _ := s.saldoRepo.FindByCardNumber(context.Background(), s.customerCardNumber)
	s.Equal(int64(850000), customerSaldo.TotalBalance)
}

func (s *WithdrawHandlerTestSuite) Test5_TrashedWithdraw() {
	s.Require().NotZero(s.withdrawID)

	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/withdraw-command/trashed/%d", s.withdrawID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *WithdrawHandlerTestSuite) Test6_RestoreWithdraw() {
	s.Require().NotZero(s.withdrawID)

	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/withdraw-command/restore/%d", s.withdrawID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *WithdrawHandlerTestSuite) Test7_PermanentDeleteWithdraw() {
	s.Require().NotZero(s.withdrawID)

	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/withdraw-command/permanent/%d", s.withdrawID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *WithdrawHandlerTestSuite) Test8_WithdrawStats_Amount() {
	ctx := context.Background()
	now := time.Now()

	err := s.chConn.Exec(ctx, "TRUNCATE TABLE withdraw_events")
	s.Require().NoError(err)

	seedSQL := `INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, card_type, card_provider, amount, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	// Success data
	err = s.chConn.Exec(ctx, seedSQL, 1, "WD001", s.customerCardNumber, "debit", "visa", 1000, "success", now)
	s.Require().NoError(err)
	// Failed data
	err = s.chConn.Exec(ctx, seedSQL, 2, "WD002", s.customerCardNumber, "debit", "visa", 2000, "failed", now)
	s.Require().NoError(err)

	s.Run("MonthlyAmount", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/amount/monthly?year=%d", now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Equal("success", resp["status"])
		s.NotEmpty(resp["data"])
	})

	s.Run("YearlyAmount", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/amount/yearly?year=%d", now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("MonthlyAmountByCard", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/amount/monthly/%s?year=%d", s.customerCardNumber, now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("YearlyAmountByCard", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/amount/yearly/%s?year=%d", s.customerCardNumber, now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})
}

func (s *WithdrawHandlerTestSuite) Test9_WithdrawStats_Status() {
	now := time.Now()

	s.Run("MonthlySuccess", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/status/monthly/success?year=%d&month=%d", now.Year(), now.Month()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("YearlyFailed", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/status/yearly/failed?year=%d", now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("MonthlySuccessByCard", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/status/monthly/success/%s?year=%d&month=%d", s.customerCardNumber, now.Year(), now.Month()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("YearlyFailedByCard", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/withdraw/stats/status/yearly/failed/%s?year=%d", s.customerCardNumber, now.Year()), nil)
		s.router.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	})
}

func (s *WithdrawHandlerTestSuite) Test10_BulkOperations() {
	// Restore All
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/withdraw-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)

	// Delete All Permanent
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/withdraw-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestWithdrawHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawHandlerTestSuite))
}
