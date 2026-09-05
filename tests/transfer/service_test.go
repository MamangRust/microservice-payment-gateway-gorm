package transfer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
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

type TransferServiceTestSuite struct {
	suite.Suite
	ts              *tests.TestSuite
	db              *gorm.DB
	redisClient     redis.UniversalClient
	transferService service.Service
	userRepo        user_repo.UserCommandRepository
	cardRepo        card_repo.Repositories
	saldoRepo       saldo_repo.Repositories
	senderCard      string
	receiverCard    string
}

func (s *TransferServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "transfer"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	saldoAdapter := &transferSaldoRepoAdapter{saldoRepo: s.saldoRepo}
	cardAdapter := &transferCardRepoAdapter{cardRepo: s.cardRepo}
	transferRepos := repository.NewRepositories(gormDB, saldoAdapter, cardAdapter)

	s.transferService = service.NewService(&service.Deps{
		Kafka: nil, Repositories: transferRepos, SaldoAdapter: s.ts.SaldoAdapter,
		CardAdapter: s.ts.CardAdapter, Logger: log, Cache: cacheStore,
	})

	ctx := context.Background()
	sender, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Sender", LastName: "Svc", Email: fmt.Sprintf("sender.svc-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	sCard, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(sender.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "111", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCard = sCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: s.senderCard, TotalBalance: 1000000})
	s.Require().NoError(err)

	receiver, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Receiver", LastName: "Svc", Email: fmt.Sprintf("receiver.svc-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	rCard, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(receiver.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "222", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCard = rCard.CardNumber
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: s.receiverCard, TotalBalance: 0})
	s.Require().NoError(err)
}

func (s *TransferServiceTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *TransferServiceTestSuite) TestCreateTransfer() {
	txn, err := s.transferService.CreateTransaction(context.Background(), &requests.CreateTransferRequest{
		TransferFrom: s.senderCard, TransferTo: s.receiverCard, TransferAmount: 50000,
	})
	s.NoError(err)
	s.NotNil(txn)
	s.Equal(int64(50000), txn.TransferAmount)
}

func TestTransferServiceSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransferServiceTestSuite))
}
