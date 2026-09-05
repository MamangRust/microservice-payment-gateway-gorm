package user_test

import (
	"context"
	"fmt"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ts   *tests.TestSuite
	db   *gorm.DB
	repo repository.UserCommandRepository
	userID int
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.repo = repository.NewUserCommandRepository(gormDB)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *UserRepositoryTestSuite) Test1_CreateUser() {
	req := &requests.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     fmt.Sprintf("john.doe.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	}

	user, err := s.repo.CreateUser(context.Background(), req)
	s.NoError(err)
	s.NotNil(user)
	s.Equal(req.FirstName, user.Firstname)
	s.Equal(req.Email, user.Email)
	s.userID = int(user.UserID)
}

func (s *UserRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	gormDB, _ := s.ts.GormDB()
	queryRepo := repository.NewUserQueryRepository(gormDB)

	found, err := queryRepo.FindById(ctx, s.userID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.userID, int(found.UserID))
}

func TestUserRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserRepositoryTestSuite))
}
