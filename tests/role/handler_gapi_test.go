package role_test

import (
	"context"
	"net"
	"testing"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleGapiTestSuite struct {
	suite.Suite
	ts            *tests.TestSuite
	db            *gorm.DB
	grpcServer    *grpc.Server
	commandClient pb.RoleCommandServiceClient
	queryClient   pb.RoleQueryServiceClient
	conn          *grpc.ClientConn
	roleID        int
}

func (s *RoleGapiTestSuite) SetupSuite() {
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

	roleService := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})

	roleHandler := handler.NewHandler(roleService)
	server := grpc.NewServer()
	pb.RegisterRoleCommandServiceServer(server, roleHandler.RoleCommand)
	pb.RegisterRoleQueryServiceServer(server, roleHandler.RoleQuery)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	go func() {
		_ = server.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.commandClient = pb.NewRoleCommandServiceClient(conn)
	s.queryClient = pb.NewRoleQueryServiceClient(conn)
}

func (s *RoleGapiTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.ts.Teardown()
}

func (s *RoleGapiTestSuite) Test1_CreateRole() {
	ctx := context.Background()
	req := &pb.CreateRoleRequest{
		Name: "Gapi Role",
	}

	res, err := s.commandClient.CreateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.Data.Name)
	s.roleID = int(res.Data.Id)
}

func (s *RoleGapiTestSuite) Test2_FindById() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	req := &pb.FindByIdRoleRequest{
		RoleId: int32(s.roleID),
	}

	res, err := s.queryClient.FindByIdRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(int32(s.roleID), res.Data.Id)
}

func (s *RoleGapiTestSuite) Test3_UpdateRole() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	req := &pb.UpdateRoleRequest{
		Id:   int32(s.roleID),
		Name: "Updated Gapi Role",
	}

	res, err := s.commandClient.UpdateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Gapi Role", res.Data.Name)
}

func (s *RoleGapiTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	req := &pb.FindByIdRoleRequest{
		RoleId: int32(s.roleID),
	}

	_, err := s.commandClient.TrashedRole(ctx, req)
	s.NoError(err)

	_, err = s.commandClient.RestoreRole(ctx, req)
	s.NoError(err)
}

func (s *RoleGapiTestSuite) Test5_DeleteRolePermanent() {
	s.Require().NotZero(s.roleID)
	ctx := context.Background()

	req := &pb.FindByIdRoleRequest{
		RoleId: int32(s.roleID),
	}

	_, err := s.commandClient.TrashedRole(ctx, req)
	s.NoError(err)

	_, err = s.commandClient.DeleteRolePermanent(ctx, req)
	s.NoError(err)
}

func (s *RoleGapiTestSuite) Test6_BulkOperations() {
	ctx := context.Background()

	// Restore All
	_, err := s.commandClient.RestoreAllRole(ctx, &emptypb.Empty{})
	s.NoError(err)

	// Delete All Permanent		_, err = s.commandClient.DeleteAllRolePermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestRoleGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleGapiTestSuite))
}
