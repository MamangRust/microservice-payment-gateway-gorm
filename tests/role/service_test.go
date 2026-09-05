package role_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type RoleServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	roleService *service.Service
	roleID      int
}

func (s *RoleServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	repos := repository.NewRepositories(gormDB)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	s.roleService = service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})
}

func (s *RoleServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *RoleServiceTestSuite) Test1_CreateRole() {
	ctx := context.Background()
	req := &requests.CreateRoleRequest{
		Name: "Service Role",
	}

	res, err := s.roleService.RoleCommand.CreateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.RoleName)
	s.roleID = int(res.RoleID)
}

func (s *RoleServiceTestSuite) Test2_FindById() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	found, err := s.roleService.RoleQuery.FindById(ctx, s.roleID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.roleID, int(found.RoleID))
}

func (s *RoleServiceTestSuite) Test3_UpdateRole() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	req := &requests.UpdateRoleRequest{
		ID:   &s.roleID,
		Name: "Updated Service Role",
	}

	res, err := s.roleService.RoleCommand.UpdateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Service Role", res.RoleName)
}

func (s *RoleServiceTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	// Trash
	_, err := s.roleService.RoleCommand.TrashedRole(ctx, s.roleID)
	s.NoError(err)

	// Restore
	_, err = s.roleService.RoleCommand.RestoreRole(ctx, s.roleID)
	s.NoError(err)
}

func (s *RoleServiceTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	_, err := s.roleService.RoleCommand.TrashedRole(ctx, s.roleID)
	s.NoError(err)

	success, err := s.roleService.RoleCommand.DeleteRolePermanent(ctx, s.roleID)
	s.NoError(err)
	s.True(success)
}

func (s *RoleServiceTestSuite) Test6_BulkOperations() {
	ctx := context.Background()

	// Restore All
	success, err := s.roleService.RoleCommand.RestoreAllRole(ctx)
	s.NoError(err)
	s.True(success)

	// Delete All Permanent
	success, err = s.roleService.RoleCommand.DeleteAllRolePermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestRoleServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleServiceTestSuite))
}
