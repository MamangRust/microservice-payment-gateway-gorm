package merchant_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
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

type MerchantServiceTestSuite struct {
	suite.Suite
	ts              *tests.TestSuite
	db              *gorm.DB
	redisClient     redis.UniversalClient
	merchantService service.Service
	userRepo        user_repo.UserCommandRepository
	userID          int
	merchantID      int
}

func (s *MerchantServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "merchant"))
	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB
	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)
	repos := repository.NewRepositories(gormDB, nil)
	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	s.merchantService = service.NewService(&service.Deps{
		Kafka: nil, Repositories: repos, UserAdapter: s.ts.UserAdapter, Logger: log, Cache: cacheStore,
	})
	user, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant", LastName: "ServiceOwner", Email: "merchant.service.owner@example.com", Password: "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *MerchantServiceTestSuite) TearDownSuite() {
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *MerchantServiceTestSuite) Test1_CreateMerchant() {
	merchant, err := s.merchantService.MerchantCommandService().CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		Name: "Service Merchant", UserID: s.userID,
	})
	s.NoError(err)
	s.NotNil(merchant)
	s.Equal("Service Merchant", merchant.Name)
	s.merchantID = int(merchant.MerchantID)
}

func (s *MerchantServiceTestSuite) Test2_FindMerchantById() {
	s.Require().NotZero(s.merchantID)
	found, err := s.merchantService.MerchantQueryService().FindById(context.Background(), s.merchantID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.merchantID, int(found.MerchantID))
}

func (s *MerchantServiceTestSuite) Test3_UpdateMerchant() {
	s.Require().NotZero(s.merchantID)
	updated, err := s.merchantService.MerchantCommandService().UpdateMerchant(context.Background(), &requests.UpdateMerchantRequest{
		MerchantID: &s.merchantID, Name: "Updated Service Merchant", UserID: s.userID, Status: "active",
	})
	s.NoError(err)
	s.NotNil(updated)
	s.Equal("Updated Service Merchant", updated.Name)
}

func (s *MerchantServiceTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.merchantID)
	_, err := s.merchantService.MerchantCommandService().TrashedMerchant(context.Background(), s.merchantID)
	s.NoError(err)
	_, err = s.merchantService.MerchantCommandService().RestoreMerchant(context.Background(), s.merchantID)
	s.NoError(err)
}

func (s *MerchantServiceTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.merchantID)
	success, err := s.merchantService.MerchantCommandService().DeleteMerchantPermanent(context.Background(), s.merchantID)
	s.NoError(err)
	s.True(success)
}

func (s *MerchantServiceTestSuite) Test7_BulkOperations() {
	success, err := s.merchantService.MerchantCommandService().RestoreAllMerchant(context.Background())
	s.NoError(err)
	s.True(success)
	success, err = s.merchantService.MerchantCommandService().DeleteAllMerchantPermanent(context.Background())
	s.NoError(err)
	s.True(success)
}

func TestMerchantServiceSuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(MerchantServiceTestSuite))
}
