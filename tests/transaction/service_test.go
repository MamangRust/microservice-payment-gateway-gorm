package transaction_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type TransactionServiceTestSuite struct {
	suite.Suite
	ts                 *tests.TestSuite
	db                 *gorm.DB
	redisClient        redis.UniversalClient
	transactionService service.Service
	userRepo           user_repo.UserCommandRepository
	cardRepo           card_repo.Repositories
	saldoRepo          saldo_repo.Repositories
	merchantRepo       merchant_repo.Repositories
	customerCardNumber string
	merchantID         int
	merchantApiKey     string
}

func (s *TransactionServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "saldo", "transaction"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil)
	s.merchantRepo = merchant_repo.NewRepositories(gormDB, nil)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	cardRepoWrapper := &transactionCardRepo{
		query: s.cardRepo.CardQuery, command: s.cardRepo.CardCommand,
	}
	transactionRepos := repository.NewRepositories(gormDB, s.saldoRepo, cardRepoWrapper, s.merchantRepo)
	s.transactionService = service.NewService(&service.Deps{
		Kafka: nil, Repositories: transactionRepos, MerchantAdapter: s.ts.MerchantAdapter,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log, Cache: cacheStore,
	})

	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Service", LastName: "Owner", Email: fmt.Sprintf("svc-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = card.CardNumber
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 1000000})
	s.Require().NoError(err)
	owner, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Merch", LastName: "Owner", Email: fmt.Sprintf("merch-owner-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	merchant, err := s.merchantRepo.CreateMerchant(ctx, &requests.CreateMerchantRequest{
		Name: "Service Merchant", UserID: int(owner.UserID),
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	s.merchantApiKey = merchant.ApiKey

	mCard, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(owner.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "321", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: mCard.CardNumber, TotalBalance: 0})
	s.Require().NoError(err)
}

func (s *TransactionServiceTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *TransactionServiceTestSuite) Test1_CreateTransaction() {
	merchantID := s.merchantID
	txn, err := s.transactionService.Create(context.Background(), s.merchantApiKey, &requests.CreateTransactionRequest{
		CardNumber: s.customerCardNumber, Amount: 100000, MerchantID: &merchantID,
		PaymentMethod: "debit", TransactionTime: time.Now(),
	})
	s.NoError(err)
	s.NotNil(txn)
	s.Equal(int64(100000), txn.Amount)
}

func TestTransactionServiceSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransactionServiceTestSuite))
}
