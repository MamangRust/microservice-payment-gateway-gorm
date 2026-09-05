package withdraw_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	withdraw_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type WithdrawServiceTestSuite struct {
	suite.Suite
	ts              *tests.TestSuite
	redisClient     redis.UniversalClient
	withdrawService service.Service
	userRepo        user_repo.UserCommandRepository
	cardRepo        card_repo.CardCommandRepository
	saldoRepo       saldo_repo.Repositories
	withdrawRepos   withdraw_repo.Repositories
	withdrawID      int32
	cardNumber      string
}

func (s *WithdrawServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "withdraw"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	// Create individual repositories from their respective modules
	userCommandRepo := user_repo.NewUserCommandRepository(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil)

	repos := withdraw_repo.NewRepositories(gormDB, cardRepos.CardQuery, saldoRepos)
	s.userRepo = userCommandRepo
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos
	s.withdrawRepos = repos

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	_ = obs
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	s.withdrawService = service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     repos,
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: nil,
	})

	// Seed User, Card and Saldo
	ctx := context.Background()
	user, _ := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Withdraw",
		LastName:  "User",
		Email:     "withdraw.service@example.com",
		Password:  "password123",
	})
	card, _ := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       int(user.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(2, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.cardNumber = card.CardNumber
	s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 500000,
	})
}

func (s *WithdrawServiceTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
}

func (s *WithdrawServiceTestSuite) Test1_Create() {
	req := &requests.CreateWithdrawRequest{
		CardNumber:     s.cardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
	}
	withdraw, err := s.withdrawService.Create(context.Background(), req)
	s.NoError(err)
	s.NotNil(withdraw)
	s.Equal(int64(req.WithdrawAmount), withdraw.WithdrawAmount)
	s.withdrawID = withdraw.WithdrawID
}

func (s *WithdrawServiceTestSuite) Test2_FindById() {
	s.Require().NotZero(s.withdrawID)

	found, err := s.withdrawService.FindById(context.Background(), int(s.withdrawID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.withdrawID, found.WithdrawID)
}

func (s *WithdrawServiceTestSuite) Test3_FindAll() {
	req := &requests.FindAllWithdraws{
		Page:     1,
		PageSize: 10,
	}
	withdraws, total, err := s.withdrawService.FindAll(context.Background(), req)
	s.NoError(err)
	s.NotNil(withdraws)
	s.NotZero(*total)
}

func (s *WithdrawServiceTestSuite) Test4_Update() {
	s.Require().NotZero(s.withdrawID)

	id := int(s.withdrawID)
	req := &requests.UpdateWithdrawRequest{
		WithdrawID:     &id,
		CardNumber:     s.cardNumber,
		WithdrawAmount: 150000,
		WithdrawTime:   time.Now(),
	}
	updated, err := s.withdrawService.Update(context.Background(), req)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int64(req.WithdrawAmount), updated.WithdrawAmount)
}

func (s *WithdrawServiceTestSuite) Test5_Trashed() {
	s.Require().NotZero(s.withdrawID)

	withdraw, err := s.withdrawService.TrashedWithdraw(context.Background(), int(s.withdrawID))
	s.NoError(err)
	s.NotNil(withdraw)
	s.NotNil(withdraw.DeletedAt)
}

func (s *WithdrawServiceTestSuite) Test6_Restore() {
	s.Require().NotZero(s.withdrawID)

	withdraw, err := s.withdrawService.RestoreWithdraw(context.Background(), int(s.withdrawID))
	s.NoError(err)
	s.NotNil(withdraw)
	s.Nil(withdraw.DeletedAt)
}

func (s *WithdrawServiceTestSuite) Test7_DeletePermanent() {
	s.Require().NotZero(s.withdrawID)

	success, err := s.withdrawService.DeleteWithdrawPermanent(context.Background(), int(s.withdrawID))
	s.NoError(err)
	s.True(success)
}

func (s *WithdrawServiceTestSuite) Test8_BulkOperations() {
	ctx := context.Background()

	// Restore All
	success, err := s.withdrawService.RestoreAllWithdraw(ctx)
	s.NoError(err)
	s.True(success)

	// Delete All Permanent
	success, err = s.withdrawService.DeleteAllWithdrawPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *WithdrawServiceTestSuite) Test9_Idempotency_SameKeyReplaysWithoutDoubleDebit() {
	ctx := context.Background()

	// Fresh card with a known balance so balance assertions stay deterministic.
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Idem", LastName: "Withdraw", Email: "idem.withdraw@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       int(user.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "777",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 300000})
	s.Require().NoError(err)

	req := &requests.CreateWithdrawRequest{
		CardNumber:     card.CardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
		IdempotencyKey: "withdraw-idem-key-1",
	}

	// First call executes and debits the balance once.
	first, err := s.withdrawService.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(first)

	// Retry with the exact same key + payload must replay, not re-execute.
	replay, err := s.withdrawService.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(replay)
	s.Equal(first.WithdrawID, replay.WithdrawID, "replay must return the original record")

	bal, err := s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(200000), bal.TotalBalance, "balance must be debited exactly once")

	// Same key with a different payload must be rejected with a conflict.
	conflictReq := &requests.CreateWithdrawRequest{
		CardNumber:     card.CardNumber,
		WithdrawAmount: 150000,
		WithdrawTime:   time.Now(),
		IdempotencyKey: "withdraw-idem-key-1",
	}
	_, err = s.withdrawService.Create(ctx, conflictReq)
	s.Require().Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr), "conflict must surface as an AppError")
	s.Equal(http.StatusConflict, appErr.Code)

	// Balance is still debited exactly once after the rejected retry.
	bal, err = s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(200000), bal.TotalBalance)
}

func (s *WithdrawServiceTestSuite) Test10_DailyWithdrawalLimit() {
	ctx := context.Background()

	// Fresh card so today's sum starts at zero, and a service instance with an
	// explicit daily limit (WITHDRAW_DAILY_LIMIT = 150000).
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Daily", LastName: "Limit", Email: "daily.limit@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       int(user.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "888",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 500000})
	s.Require().NoError(err)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	obs, _ := observability.NewObservability("test", log)
	_ = obs
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	svc := service.NewService(&service.Deps{
		Kafka:                nil,
		Repositories:         s.withdrawRepos,
		CardAdapter:          s.ts.CardAdapter,
		SaldoAdapter:         s.ts.SaldoAdapter,
		Logger:               log,
		Cache:                cacheStore,
		AISecurityClient:     nil,
		DailyWithdrawalLimit: 150000,
	})

	// First withdrawal is within the limit.
	first, err := svc.Create(ctx, &requests.CreateWithdrawRequest{
		CardNumber:     card.CardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(first)

	// Second withdrawal pushes today's total past the limit: rejected with 400.
	overLimit, err := svc.Create(ctx, &requests.CreateWithdrawRequest{
		CardNumber:     card.CardNumber,
		WithdrawAmount: 60000,
		WithdrawTime:   time.Now(),
	})
	s.Require().Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr), "over-limit must surface as an AppError")
	s.Equal(http.StatusBadRequest, appErr.Code, "over-limit must be a BadRequest")
	s.Nil(overLimit)

	// Balance must be untouched by the rejected withdrawal (only 100000 debited).
	bal, err := s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(400000), bal.TotalBalance, "rejected withdrawal must not change the balance")
}

func TestWithdrawServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawServiceTestSuite))
}
