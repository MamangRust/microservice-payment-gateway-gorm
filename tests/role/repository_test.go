package role_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	ts   *tests.TestSuite
	db   *gorm.DB
	repo repository.Repositories
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.repo = repository.NewRepositories(gormDB)
}

func (s *RoleRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *RoleRepositoryTestSuite) createSeedRole() (*models.Role, error) {
	return s.repo.CreateRole(context.Background(), &requests.CreateRoleRequest{
		Name: fmt.Sprintf("Test Role-%d", time.Now().UnixNano()),
	})
}

func (s *RoleRepositoryTestSuite) TestCreateRole() {
	ctx := context.Background()
	req := &requests.CreateRoleRequest{
		Name: "Test Role",
	}

	res, err := s.repo.CreateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.RoleName)
}

func (s *RoleRepositoryTestSuite) TestFindAllRoles() {
	_, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindAllRoles(ctx, &requests.FindAllRoles{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *RoleRepositoryTestSuite) TestFindById() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindById(ctx, int(role.RoleID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(role.RoleID, found.RoleID)
}

func (s *RoleRepositoryTestSuite) TestFindByActiveRole() {
	_, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindByActiveRole(ctx, &requests.FindAllRoles{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *RoleRepositoryTestSuite) TestFindByTrashedRole() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedRole(ctx, int(role.RoleID))
	s.Require().NoError(err)

	res, err := s.repo.FindByTrashedRole(ctx, &requests.FindAllRoles{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *RoleRepositoryTestSuite) TestUpdateRole() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	id := int(role.RoleID)
	req := &requests.UpdateRoleRequest{
		ID:   &id,
		Name: "Updated Role",
	}

	res, err := s.repo.UpdateRole(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Role", res.RoleName)
}

func (s *RoleRepositoryTestSuite) TestTrashRole() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.TrashedRole(ctx, int(role.RoleID))
	s.NoError(err)
	s.NotNil(trashed)
	s.NotNil(trashed.DeletedAt)
}

func (s *RoleRepositoryTestSuite) TestRestoreRole() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedRole(ctx, int(role.RoleID))
	s.Require().NoError(err)

	restored, err := s.repo.RestoreRole(ctx, int(role.RoleID))
	s.NoError(err)
	s.NotNil(restored)
	s.Nil(restored.DeletedAt)
}

func (s *RoleRepositoryTestSuite) TestDeleteRolePermanent() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedRole(ctx, int(role.RoleID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteRolePermanent(ctx, int(role.RoleID))
	s.NoError(err)
	s.True(success)

	_, err = s.repo.FindById(ctx, int(role.RoleID))
	s.Error(err)
}

func (s *RoleRepositoryTestSuite) TestRestoreAllRole() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedRole(ctx, int(role.RoleID))
	s.Require().NoError(err)

	success, err := s.repo.RestoreAllRole(ctx)
	s.NoError(err)
	s.True(success)

	found, err := s.repo.FindById(ctx, int(role.RoleID))
	s.NoError(err)
	s.NotNil(found)
}

func (s *RoleRepositoryTestSuite) TestDeleteAllRolePermanent() {
	role, err := s.createSeedRole()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedRole(ctx, int(role.RoleID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteAllRolePermanent(ctx)
	s.NoError(err)
	s.True(success)

	_, err = s.repo.FindById(ctx, int(role.RoleID))
	s.Error(err)
}

func TestRoleRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleRepositoryTestSuite))
}
