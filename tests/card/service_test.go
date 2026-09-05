package card_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
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

type CardServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	service     service.Service
	userRepo    user_repo.Repositories
	redisClient redis.UniversalClient
	userID      int
	cardID      int
}

func (s *CardServiceTestSuite) SetupSuite() {
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

	repos := repository.NewRepositories(gormDB, nil)
	s.userRepo = user_repo.NewRepositories(gormDB)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	s.service = service.NewService(&service.Deps{
		Repositories: repos,
		UserAdapter:  s.ts.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	// Create a user for card ownership
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Card",
		LastName:  "Service",
		Email:     "card.service@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *CardServiceTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *CardServiceTestSuite) Test1_CreateCard() {
	ctx := context.Background()
	req := &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	}

	res, err := s.service.CreateCard(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.NotEmpty(res.CardNumber)
	s.cardID = int(res.CardID)
}

func (s *CardServiceTestSuite) Test2_FindById() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()

	found, err := s.service.FindById(ctx, s.cardID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.cardID), found.CardID)
}

func (s *CardServiceTestSuite) Test3_UpdateCard() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()

	req := &requests.UpdateCardRequest{
		CardID:       s.cardID,
		UserID:       s.userID,
		CardType:     "credit",
		ExpireDate:   time.Now().AddDate(6, 0, 0),
		CVV:          "456",
		CardProvider: "MasterCard",
	}

	res, err := s.service.UpdateCard(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("credit", res.CardType)
}

func (s *CardServiceTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()

	trashed, err := s.service.TrashedCard(ctx, s.cardID)
	s.NoError(err)
	s.NotNil(trashed)

	restored, err := s.service.RestoreCard(ctx, s.cardID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *CardServiceTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.cardID)
	ctx := context.Background()

	trashed, _ := s.service.TrashedCard(ctx, s.cardID)
	s.NotNil(trashed)

	success, err := s.service.DeleteCardPermanent(ctx, s.cardID)
	s.NoError(err)
	s.True(success)
}

func (s *CardServiceTestSuite) Test6_BulkOperations() {
	ctx := context.Background()

	// Restore All
	success, err := s.service.RestoreAllCard(ctx)
	s.NoError(err)
	s.True(success)

	// Delete All Permanent
	success, err = s.service.DeleteAllCardPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestCardServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardServiceTestSuite))
}
