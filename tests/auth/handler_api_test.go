package auth_test

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/auth"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	authhandler "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/handler/auth"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	redisClient redis.UniversalClient
	server      *echo.Echo
	email       string
	password    string
	accessToken string
	userID      int
}

func (s *AuthHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	repos := repository.NewRepositories(
		gormDB,
		s.ts.UserQueryClient,
		s.ts.UserCommandClient,
		s.ts.RoleQueryClient,
		s.ts.UserRoleClient,
	)

	tokenManager, _ := auth.NewManager("mysecret")
	hasher := hash.NewHashingPassword()

	svc := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       s.ts.Logger,
		Cache:        s.ts.CacheStore,
		Token:        tokenManager,
		Hash:         hasher,
		Kafka:        nil,
	})

	h := handler.NewAuthHandleGrpc(svc, s.ts.Logger)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.server = echo.New()

	apiHandler := errors.NewApiHandler(s.ts.Observability, s.ts.Logger)

	// Auth bypass middleware for /api/auth/me — the handler reads c.Get("user_id")
	// (set by the production JWT middleware), so set that key, not "userId".
	s.server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if s.userID != 0 {
				c.Set("user_id", strconv.Itoa(s.userID))
			}
			return next(c)
		}
	})

	authhandler.RegisterAuthHandler(&authhandler.DepsAuth{
		Client:     conn,
		E:          s.server,
		Logger:     s.ts.Logger,
		Cache:      s.ts.CacheStore,
		ApiHandler: apiHandler,
	})

	s.email = "auth.handler.api.test@example.com"
	s.password = "password123"

	// Seed ROLE_ADMIN
	_, _ = s.ts.SeedRole(gormDB, "ROLE_ADMIN")
}

func (s *AuthHandlerApiTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	s.ts.Teardown()
}

func (s *AuthHandlerApiTestSuite) Test1_Register() {
	body := map[string]string{
		"firstname":        "Auth",
		"lastname":         "API",
		"email":            s.email,
		"password":         s.password,
		"confirm_password": s.password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.userID = int(data["id"].(float64))
}

func (s *AuthHandlerApiTestSuite) Test2_Login() {
	body := map[string]string{
		"email":    s.email,
		"password": s.password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)

	data := res["data"].(map[string]interface{})
	s.accessToken = data["access_token"].(string)
}

func (s *AuthHandlerApiTestSuite) Test4_LoginLockout() {
	email := "locked.api@example.com"
	password := "wrongpassword"

	// Register user first
	regBody := map[string]string{
		"firstname":        "Locked",
		"lastname":         "API",
		"email":            email,
		"password":         "correctpassword",
		"confirm_password": "correctpassword",
	}
	jsonRegBody, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(jsonRegBody))
	regReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	regRec := httptest.NewRecorder()
	s.server.ServeHTTP(regRec, regReq)
	s.Equal(http.StatusCreated, regRec.Code)

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonLoginBody, _ := json.Marshal(loginBody)

	// Fail login 5 times (total 5)
	for i := 0; i < 5; i++ {
		loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
		loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec := httptest.NewRecorder()
		s.server.ServeHTTP(loginRec, loginReq)
		s.Equal(http.StatusUnauthorized, loginRec.Code)
	}

	// 6th attempt should return 403 Forbidden (ErrAccountLocked)
	lockedReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
	lockedReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	lockedRec := httptest.NewRecorder()
	s.server.ServeHTTP(lockedRec, lockedReq)
	s.Equal(http.StatusForbidden, lockedRec.Code)
}

func (s *AuthHandlerApiTestSuite) Test3_GetMe() {
	s.Require().NotZero(s.userID)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal(s.email, data["email"])
}

func TestAuthHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerApiTestSuite))
}
