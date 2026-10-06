package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	api "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/user"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserHandlerTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	grpcServer  *grpc.Server
	client      pb.UserCommandServiceClient
	conn        *grpc.ClientConn
	router      *echo.Echo
	userID      int
	userEmail   string
}

func (s *UserHandlerTestSuite) SetupSuite() {
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

	repos := repository.NewRepositories(&repository.Deps{
		Db:              gormDB,
		RoleQueryClient: s.ts.RoleClient,
		UserRoleClient:  s.ts.UserRoleClient,
		Guard:           repository.GuardOptions{},
	})

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	hasher := hash.NewHashingPassword()
	obs, _ := observability.NewObservability("test", log)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	userService := service.NewService(&service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		Hash:         hasher,
		Logger:       log,
	})

	// Start gRPC Server
	userHandler := gapi.NewHandler(userService)
	server := grpc.NewServer()
	pb.RegisterUserQueryServiceServer(server, userHandler)
	pb.RegisterUserCommandServiceServer(server, userHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)

	go func() {
		_ = server.Serve(lis)
	}()

	// Create gRPC Client
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pb.NewUserCommandServiceClient(conn)

	// Setup Echo
	s.router = echo.New()
	apiErrorHandler := app_errors.NewApiHandler(obs, log)

	api.RegisterUserHandler(&api.DepsUser{
		Client:     conn,
		E:          s.router,
		Logger:     log,
		Cache:      cacheStore,
		ApiHandler: apiErrorHandler,
	})
}

func (s *UserHandlerTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.ts.Teardown()
}

func (s *UserHandlerTestSuite) Test1_CreateUser() {
	s.userEmail = fmt.Sprintf("handler.user.%d@example.com", time.Now().UnixNano())
	req := requests.CreateUserRequest{
		FirstName:       "Handler",
		LastName:        "User",
		Email:           s.userEmail,
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest(http.MethodPost, "/api/user-command/create", bytes.NewBuffer(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, request)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	var createRes map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createRes)
	userData := createRes["data"].(map[string]interface{})
	s.userID = int(userData["id"].(float64))
}

func (s *UserHandlerTestSuite) Test2_FindUserById() {
	s.Require().NotZero(s.userID)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/user-query/%d", s.userID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *UserHandlerTestSuite) Test3_UpdateUser() {
	s.Require().NotZero(s.userID)

	updateReq := requests.UpdateUserRequest{
		FirstName:       "Updated",
		LastName:        "User",
		Email:           s.userEmail,
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	updateBody, _ := json.Marshal(updateReq)
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/user-command/update/%d", s.userID), bytes.NewBuffer(updateBody))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *UserHandlerTestSuite) Test4_PermanentDeleteUser() {
	s.Require().NotZero(s.userID)

	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/user-command/permanent/%d", s.userID), nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, request)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *UserHandlerTestSuite) Test5_BulkOperations() {
	// Restore All
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/user-command/restore/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)

	// Delete All Permanent
	rec = httptest.NewRecorder()
	httpReq = httptest.NewRequest(http.MethodPost, "/api/user-command/permanent/all", nil)
	s.router.ServeHTTP(rec, httpReq)
	s.Equal(http.StatusOK, rec.Code)
}

func TestUserHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserHandlerTestSuite))
}
