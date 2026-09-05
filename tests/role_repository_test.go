package tests

import (
	"testing"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RoleRepositoryTestSuite struct {
	suite.Suite
	ts   *TestSuite
	repo repository.Repositories
}

func TestRoleRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(RoleRepositoryTestSuite))
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	ts, err := SetupTestSuite()
	require.NoError(s.T(), err)
	s.ts = ts

	err = s.ts.RunServiceMigrations("role")
	require.NoError(s.T(), err)

	gormDB, err := s.ts.GormDB()
	require.NoError(s.T(), err)

	s.repo = repository.NewRepositories(gormDB)
}

func (s *RoleRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *RoleRepositoryTestSuite) TestCreateRole() {
	req := &requests.CreateRoleRequest{
		Name: "TEST_ROLE",
	}

	role, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), role.RoleID)
	assert.Equal(s.T(), "TEST_ROLE", role.RoleName)
}

func (s *RoleRepositoryTestSuite) TestFindById() {
	req := &requests.CreateRoleRequest{
		Name: "FIND_BY_ID_ROLE",
	}
	created, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	role, err := s.repo.FindById(s.ts.Ctx, int(created.RoleID))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "FIND_BY_ID_ROLE", role.RoleName)
}

func (s *RoleRepositoryTestSuite) TestFindByName() {
	req := &requests.CreateRoleRequest{
		Name: "FIND_BY_NAME_ROLE",
	}
	_, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	role, err := s.repo.FindByName(s.ts.Ctx, "FIND_BY_NAME_ROLE")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "FIND_BY_NAME_ROLE", role.RoleName)
}

func (s *RoleRepositoryTestSuite) TestFindAllRoles() {
	req := &requests.FindAllRoles{
		Page:     1,
		PageSize: 10,
		Search:   "",
	}

	roles, err := s.repo.FindAllRoles(s.ts.Ctx, req)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), roles)
}

func (s *RoleRepositoryTestSuite) TestUpdateRole() {
	createReq := &requests.CreateRoleRequest{
		Name: "UPDATE_ROLE_OLD",
	}
	created, err := s.repo.CreateRole(s.ts.Ctx, createReq)
	require.NoError(s.T(), err)

	id := int(created.RoleID)
	updateReq := &requests.UpdateRoleRequest{
		ID:   &id,
		Name: "UPDATE_ROLE_NEW",
	}
	updated, err := s.repo.UpdateRole(s.ts.Ctx, updateReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "UPDATE_ROLE_NEW", updated.RoleName)
}

func (s *RoleRepositoryTestSuite) TestTrashedRole() {
	req := &requests.CreateRoleRequest{
		Name: "TRASH_ROLE",
	}
	created, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	trashed, err := s.repo.TrashedRole(s.ts.Ctx, int(created.RoleID))
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), trashed.DeletedAt)
}

func (s *RoleRepositoryTestSuite) TestRestoreRole() {
	req := &requests.CreateRoleRequest{
		Name: "RESTORE_ROLE",
	}
	created, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	_, err = s.repo.TrashedRole(s.ts.Ctx, int(created.RoleID))
	require.NoError(s.T(), err)

	restored, err := s.repo.RestoreRole(s.ts.Ctx, int(created.RoleID))
	require.NoError(s.T(), err)
	assert.Nil(s.T(), restored.DeletedAt)
}

func (s *RoleRepositoryTestSuite) TestDeleteRolePermanent() {
	req := &requests.CreateRoleRequest{
		Name: "DELETE_PERM_ROLE",
	}
	created, err := s.repo.CreateRole(s.ts.Ctx, req)
	require.NoError(s.T(), err)

	deleted, err := s.repo.DeleteRolePermanent(s.ts.Ctx, int(created.RoleID))
	require.NoError(s.T(), err)
	assert.True(s.T(), deleted)
}

func (s *RoleRepositoryTestSuite) TestCreateUserRole() {
	roleReq := &requests.CreateRoleRequest{
		Name: "USER_ROLE_TEST",
	}
	role, err := s.repo.CreateRole(s.ts.Ctx, roleReq)
	require.NoError(s.T(), err)

	userRole, err := s.repo.CreateUserRole(s.ts.Ctx, 1, int(role.RoleID))
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), userRole)
}
