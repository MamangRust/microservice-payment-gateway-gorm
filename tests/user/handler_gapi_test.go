package user_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
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

type UserGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	client      pb.UserCommandServiceClient
	queryClient pb.UserQueryServiceClient
	conn        *grpc.ClientConn
	userID      int
}

func (s *UserGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	repos := repository.NewRepositories(gormDB)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	userService := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Hash:         hasher,
		Cache:        cacheStore,
	})

	// Start gRPC Server
	userHandler := gapi.NewHandler(userService)
	server := grpc.NewServer()
	pb.RegisterUserCommandServiceServer(server, userHandler)
	pb.RegisterUserQueryServiceServer(server, userHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	go func() {
		_ = server.Serve(lis)
	}()

	// Create Client
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pb.NewUserCommandServiceClient(conn)
	s.queryClient = pb.NewUserQueryServiceClient(conn)
}

func (s *UserGapiTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *UserGapiTestSuite) Test1_CreateUser() {
	ctx := context.Background()

	// 1. Create
	createReq := &pb.CreateUserRequest{
		Firstname:       "Gapi",
		Lastname:        "User",
		Email:           fmt.Sprintf("gapi.user.%d@example.com", time.Now().UnixNano()),
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	res, err := s.client.Create(ctx, createReq)
	s.NoError(err)
	s.Equal(createReq.Email, res.Data.Email)
	s.userID = int(res.Data.Id)
}

func (s *UserGapiTestSuite) Test2_FindUserById() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	findReq := &pb.FindByIdUserRequest{
		Id: int32(s.userID),
	}
	found, err := s.queryClient.FindById(ctx, findReq)
	s.NoError(err)
	s.Equal(int32(s.userID), found.Data.Id)
}

func (s *UserGapiTestSuite) Test3_BulkOperations() {
	ctx := context.Background()

	// Restore All
	_, err := s.client.RestoreAllUser(ctx, &emptypb.Empty{})
	s.NoError(err)

	// Delete All Permanent
	_, err = s.client.DeleteAllUserPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestUserGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGapiTestSuite))
}
