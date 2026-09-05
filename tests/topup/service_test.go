package topup_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	topup_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type TopupServiceTestSuite struct {
	suite.Suite
	ts           *tests.TestSuite
	db           *gorm.DB
	redisClient  redis.UniversalClient
	topupService service.Service
	userRepo     user_repo.UserCommandRepository
	cardRepo     card_repo.CardCommandRepository
	saldoRepo    saldo_repo.Repositories
	topupID      int32
	cardNumber   string
}

func (s *TopupServiceTestSuite) SetupSuite() {
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

	userRepos := user_repo.NewRepositories(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil)

	cardAdapter := &topupCardRepoAdapter{
		CardQueryRepository:   cardRepos.CardQuery,
		CardCommandRepository: cardRepos.CardCommand,
	}
	topupRepos := topup_repo.NewRepositories(gormDB, cardAdapter, saldoRepos)

	s.userRepo = userRepos.UserCommand()
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	s.topupService = service.NewService(&service.Deps{
		Kafka: nil, Cache: cacheStore, Repositories: topupRepos,
		CardAdapter: s.ts.CardAdapter, SaldoAdapter: s.ts.SaldoAdapter, Logger: log,
	})

	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Topup", LastName: "Owner", Email: "topup.service@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0),
		CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: s.cardNumber, TotalBalance: 0})
	s.Require().NoError(err)
}

func (s *TopupServiceTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *TopupServiceTestSuite) Test1_CreateTopup() {
	topup, err := s.topupService.CreateTopup(context.Background(), &requests.CreateTopupRequest{
		CardNumber: s.cardNumber, TopupAmount: 100000, TopupMethod: "visa",
	})
	s.NoError(err)
	s.NotNil(topup)
	s.Equal(int64(100000), topup.TopupAmount)
	s.topupID = topup.TopupID
}

func (s *TopupServiceTestSuite) Test2_FindById() {
	s.Require().NotZero(s.topupID)
	found, err := s.topupService.FindById(context.Background(), int(s.topupID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int64(100000), found.TopupAmount)
}

func (s *TopupServiceTestSuite) Test3_FindAll() {
	topups, total, err := s.topupService.FindAll(context.Background(), &requests.FindAllTopups{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotNil(topups)
	s.NotZero(*total)
}

func (s *TopupServiceTestSuite) Test4_UpdateTopup() {
	s.Require().NotZero(s.topupID)
	id := int(s.topupID)
	updated, err := s.topupService.UpdateTopup(context.Background(), &requests.UpdateTopupRequest{
		TopupID: &id, CardNumber: s.cardNumber, TopupAmount: 150000, TopupMethod: "visa",
	})
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int64(150000), updated.TopupAmount)
}

func (s *TopupServiceTestSuite) Test5_TrashAndRestore() {
	s.Require().NotZero(s.topupID)
	_, err := s.topupService.TrashedTopup(context.Background(), int(s.topupID))
	s.NoError(err)
	_, err = s.topupService.RestoreTopup(context.Background(), int(s.topupID))
	s.NoError(err)
}

func (s *TopupServiceTestSuite) Test6_DeletePermanent() {
	s.Require().NotZero(s.topupID)
	success, err := s.topupService.DeleteTopupPermanent(context.Background(), int(s.topupID))
	s.NoError(err)
	s.True(success)
}

func (s *TopupServiceTestSuite) Test7_BulkOperations() {
	success, err := s.topupService.RestoreAllTopup(context.Background())
	s.NoError(err)
	s.True(success)
	success, err = s.topupService.DeleteAllTopupPermanent(context.Background())
	s.NoError(err)
	s.True(success)
}

func (s *TopupServiceTestSuite) Test8_Idempotency_SameKeyReplaysWithoutDoubleCredit() {
	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Idem", LastName: "Topup", Email: "idem.topup@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0),
		CVV: "555", CardProvider: "visa",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 0})
	s.Require().NoError(err)

	req := &requests.CreateTopupRequest{
		CardNumber: card.CardNumber, TopupAmount: 50000, TopupMethod: "visa", IdempotencyKey: "topup-idem-key-1",
	}
	first, err := s.topupService.CreateTopup(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(first)

	replay, err := s.topupService.CreateTopup(ctx, req)
	s.Require().NoError(err)
	s.Equal(first.TopupID, replay.TopupID)

	bal, err := s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(50000), bal.TotalBalance)

	conflictReq := &requests.CreateTopupRequest{
		CardNumber: card.CardNumber, TopupAmount: 75000, TopupMethod: "visa", IdempotencyKey: "topup-idem-key-1",
	}
	_, err = s.topupService.CreateTopup(ctx, conflictReq)
	s.Require().Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr))
	s.Equal(http.StatusConflict, appErr.Code)

	bal, err = s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(50000), bal.TotalBalance)
}

func (s *TopupServiceTestSuite) Test9_Idempotency_ConcurrentSameKeyCreditsOnce() {
	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Idem", LastName: "Concurrent", Email: "idem.concurrent@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0),
		CVV: "999", CardProvider: "visa",
	})
	s.Require().NoError(err)
	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: 0})
	s.Require().NoError(err)

	req := &requests.CreateTopupRequest{
		CardNumber: card.CardNumber, TopupAmount: 50000, TopupMethod: "visa", IdempotencyKey: "topup-concurrent-key-1",
	}
	const workers = 5
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.topupService.CreateTopup(ctx, req)
		}()
	}
	wg.Wait()

	bal, err := s.saldoRepo.FindByCardNumber(ctx, card.CardNumber)
	s.Require().NoError(err)
	s.Equal(int64(50000), bal.TotalBalance)
}

func TestTopupServiceSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TopupServiceTestSuite))
}
