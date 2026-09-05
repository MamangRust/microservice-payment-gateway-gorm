package tests

import (
	"testing"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func intPtr(v int) *int { return &v }

type UserRepositoryTestSuite struct {
	suite.Suite
	ts     *TestSuite
	repo   repository.Repositories
}

func TestUserRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	ts, err := SetupTestSuite()
	require.NoError(s.T(), err)
	s.ts = ts

	// Run migrations in a single pass so goose versions don't collide across
	// separate runs (each RunAllMigrations call renumbers from 0001).
	err = s.ts.RunMigrations("user", "role")
	require.NoError(s.T(), err)

	// Initialize GORM DB and repository
	gormDB, err := s.ts.GormDB()
	require.NoError(s.T(), err)

	s.repo = repository.NewRepositories(gormDB)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	req := &requests.CreateUserRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}

	user, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), user.UserID)
	assert.Equal(s.T(), "John", user.Firstname)
	assert.Equal(s.T(), "Doe", user.Lastname)
	assert.Equal(s.T(), "john@example.com", user.Email)
}

func (s *UserRepositoryTestSuite) TestFindById() {
	// Create a user first
	req := &requests.CreateUserRequest{
		FirstName:       "Jane",
		LastName:        "Smith",
		Email:           "jane@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	// Find by ID
	user, err := s.repo.UserQuery().FindById(s.ts.Ctx, int(created.UserID))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Jane", user.Firstname)
	assert.Equal(s.T(), "Smith", user.Lastname)
}

func (s *UserRepositoryTestSuite) TestFindByEmail() {
	req := &requests.CreateUserRequest{
		FirstName:       "Find",
		LastName:        "ByEmail",
		Email:           "findbyemail@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	_, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	user, err := s.repo.UserQuery().FindByEmail(s.ts.Ctx, "findbyemail@example.com")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Find", user.Firstname)
}

func (s *UserRepositoryTestSuite) TestFindAllUsers() {
	req := &requests.FindAllUsers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	}

	users, err := s.repo.UserQuery().FindAllUsers(s.ts.Ctx, req)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), users)
}

func (s *UserRepositoryTestSuite) TestUpdateUser() {
	// Create a user
	createReq := &requests.CreateUserRequest{
		FirstName:       "Update",
		LastName:        "Me",
		Email:           "update@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, createReq)
	require.NoError(s.T(), err)

	// Update it
	updateReq := &requests.UpdateUserRequest{
		UserID:          intPtr(int(created.UserID)),
		FirstName:       "Updated",
		LastName:        "User",
		Email:           "updated@example.com",
		Password:        "new_hashed_password",
		ConfirmPassword: "new_hashed_password",
	}
	updated, err := s.repo.UserCommand().UpdateUser(s.ts.Ctx, updateReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated", updated.Firstname)
	assert.Equal(s.T(), "User", updated.Lastname)
}

func (s *UserRepositoryTestSuite) TestTrashedUser() {
	req := &requests.CreateUserRequest{
		FirstName:       "Trash",
		LastName:        "Me",
		Email:           "trash@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	trashed, err := s.repo.UserCommand().TrashedUser(s.ts.Ctx, int(created.UserID))
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), trashed.DeletedAt)
}

func (s *UserRepositoryTestSuite) TestRestoreUser() {
	req := &requests.CreateUserRequest{
		FirstName:       "Restore",
		LastName:        "Me",
		Email:           "restore@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	// Trash it
	_, err = s.repo.UserCommand().TrashedUser(s.ts.Ctx, int(created.UserID))
	require.NoError(s.T(), err)

	// Restore it
	restored, err := s.repo.UserCommand().RestoreUser(s.ts.Ctx, int(created.UserID))
	require.NoError(s.T(), err)
	assert.Nil(s.T(), restored.DeletedAt)
}

func (s *UserRepositoryTestSuite) TestDeleteUserPermanent() {
	req := &requests.CreateUserRequest{
		FirstName:       "Delete",
		LastName:        "Permanently",
		Email:           "deleteperm@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	deleted, err := s.repo.UserCommand().DeleteUserPermanent(s.ts.Ctx, int(created.UserID))
	require.NoError(s.T(), err)
	assert.True(s.T(), deleted)
}

func (s *UserRepositoryTestSuite) TestUpdateIsVerified() {
	req := &requests.CreateUserRequest{
		FirstName:       "Verify",
		LastName:        "Me",
		Email:           "verify@example.com",
		Password:        "hashed_password",
		ConfirmPassword: "hashed_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	updated, err := s.repo.UserCommand().UpdateIsVerified(s.ts.Ctx, int(created.UserID), true)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), updated)
}

func (s *UserRepositoryTestSuite) TestUpdatePassword() {
	req := &requests.CreateUserRequest{
		FirstName:       "Password",
		LastName:        "Update",
		Email:           "password@example.com",
		Password:        "old_password",
		ConfirmPassword: "old_password",
	}
	created, err := s.repo.UserCommand().CreateUser(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	updated, err := s.repo.UserCommand().UpdatePassword(s.ts.Ctx, int(created.UserID), "new_password")
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), updated)
}
