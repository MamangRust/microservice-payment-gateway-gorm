package transaction_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"

	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"

)

type faultInjectingSaldoAdapter struct {
	inner      adapter.SaldoAdapter
	failDebit  bool
	failCredit bool
}

func (f *faultInjectingSaldoAdapter) FindByCardNumber(ctx context.Context, cardNumber string) (*models.Saldo, error) {
	return f.inner.FindByCardNumber(ctx, cardNumber)
}

func (f *faultInjectingSaldoAdapter) UpdateSaldoBalance(ctx context.Context, req *requests.UpdateSaldoBalance) (*saldo_repo.SaldoMutationResult, error) {
	return f.inner.UpdateSaldoBalance(ctx, req)
}

func (f *faultInjectingSaldoAdapter) DebitSaldo(ctx context.Context, req *requests.DebitSaldoRequest) (*saldo_repo.SaldoMutationResult, error) {
	if f.failDebit {
		return nil, app_errors.NewServiceUnavailableError("saldo service unavailable (injected)")
	}
	return f.inner.DebitSaldo(ctx, req)
}

func (f *faultInjectingSaldoAdapter) CreditSaldo(ctx context.Context, req *requests.CreditSaldoRequest) (*saldo_repo.SaldoMutationResult, error) {
	if f.failCredit {
		return nil, app_errors.NewServiceUnavailableError("saldo service unavailable (injected)")
	}
	return f.inner.CreditSaldo(ctx, req)
}

func (f *faultInjectingSaldoAdapter) UpdateSaldoWithdraw(ctx context.Context, req *requests.UpdateSaldoWithdraw) (*saldo_repo.SaldoMutationResult, error) {
	return f.inner.UpdateSaldoWithdraw(ctx, req)
}

type TransactionFailureInjectionTestSuite struct {
	suite.Suite
	ts             *tests.TestSuite
	db             *gorm.DB
	redisClient    redis.UniversalClient
	transactionSvc service.Service
	injectedSaldo  *faultInjectingSaldoAdapter
	customerCard   string
	merchantID     int
	merchantApiKey string
}

func (s *TransactionFailureInjectionTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "saldo", "transaction"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	userRepo := user_repo.NewUserCommandRepository(gormDB)
	cardRepo := *card_repo.NewRepositories(gormDB, nil)
	merchantRepo := merchant_repo.NewRepositories(gormDB, nil)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("failure-injection-test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("failure-injection-test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	s.injectedSaldo = &faultInjectingSaldoAdapter{inner: s.ts.SaldoAdapter}

	cardRepoWrapper := &transactionCardRepo{
		query: cardRepo.CardQuery, command: cardRepo.CardCommand,
	}
	transactionRepos := repository.NewRepositories(gormDB, s.injectedSaldo, cardRepoWrapper, merchantRepo)
	s.transactionSvc = service.NewService(&service.Deps{
		Kafka: nil, Repositories: transactionRepos, MerchantAdapter: s.ts.MerchantAdapter,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.injectedSaldo, Logger: log, Cache: cacheStore,
	})

	ctx := context.Background()
	user, err := userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Fail", LastName: "Test", Email: "fail.test@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCard = card.CardNumber
	_, err = saldo_repo.NewRepositories(gormDB, nil).CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 1000000})
	s.Require().NoError(err)
	owner, err := userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Fail", LastName: "Owner", Email: "fail.owner@test.com", Password: "password123",
	})
	s.Require().NoError(err)
	merchant, err := merchantRepo.CreateMerchant(ctx, &requests.CreateMerchantRequest{
		Name: "Fail Merchant", UserID: int(owner.UserID),
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	s.merchantApiKey = merchant.ApiKey

	mCard, err := cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(owner.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "321", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	_, err = saldo_repo.NewRepositories(gormDB, nil).CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: mCard.CardNumber, TotalBalance: 0})
	s.Require().NoError(err)
}

func (s *TransactionFailureInjectionTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *TransactionFailureInjectionTestSuite) TestDebitFailureReturnsError() {
	s.injectedSaldo.failDebit = true
	defer func() { s.injectedSaldo.failDebit = false }()

	merchantID := s.merchantID
	_, err := s.transactionSvc.Create(context.Background(), s.merchantApiKey, &requests.CreateTransactionRequest{
		CardNumber: s.customerCard, Amount: 50000, MerchantID: &merchantID,
		PaymentMethod: "debit", TransactionTime: time.Now(),
	})
	s.Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr))
	s.Equal(http.StatusServiceUnavailable, appErr.Code)
}

func (s *TransactionFailureInjectionTestSuite) TestCreditFailureReturnsError() {
	s.injectedSaldo.failCredit = true
	defer func() { s.injectedSaldo.failCredit = false }()

	merchantID := s.merchantID
	_, err := s.transactionSvc.Create(context.Background(), s.merchantApiKey, &requests.CreateTransactionRequest{
		CardNumber: s.customerCard, Amount: 50000, MerchantID: &merchantID,
		PaymentMethod: "credit", TransactionTime: time.Now(),
	})
	s.Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr))
	s.Equal(http.StatusServiceUnavailable, appErr.Code)
}

func TestTransactionFailureInjectionSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransactionFailureInjectionTestSuite))
}
