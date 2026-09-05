package handler

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/stretchr/testify/require"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
)

type roleQueryServiceStub struct {
	active  []*models.Role
	trashed []*models.Role
}

func (s *roleQueryServiceStub) FindAll(context.Context, *requests.FindAllRoles) ([]*models.Role, *int, error) {
	return nil, intPointer(0), nil
}

func (s *roleQueryServiceStub) FindByActiveRole(context.Context, *requests.FindAllRoles) ([]*models.Role, *int, error) {
	total := len(s.active)
	return s.active, &total, nil
}

func (s *roleQueryServiceStub) FindByTrashedRole(context.Context, *requests.FindAllRoles) ([]*models.Role, *int, error) {
	total := len(s.trashed)
	return s.trashed, &total, nil
}

func (s *roleQueryServiceStub) FindById(context.Context, int) (*models.Role, error) {
	return nil, nil
}

func (s *roleQueryServiceStub) FindByUserId(context.Context, int) ([]*models.Role, error) {
	return nil, nil
}

func (s *roleQueryServiceStub) FindByName(context.Context, string) (*models.Role, error) {
	return nil, nil
}

func intPointer(value int) *int {
	return &value
}

func TestRoleQueryHandleGrpcFindByTrashed_NullSafeDeletedAt(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC)

	stub := &roleQueryServiceStub{
		trashed: []*models.Role{
			{RoleID: 7, RoleName: "ROLE_TRASHED", CreatedAt: &createdAt, UpdatedAt: &updatedAt, DeletedAt: &deletedAt},
			// A malformed/null row must not panic the handler or create a bogus date.
			{RoleID: 8, RoleName: "ROLE_NULL_DATE", CreatedAt: &createdAt, UpdatedAt: &updatedAt},
			nil,
		},
	}

	h := NewRoleQueryHandleGrpc(stub)
	res, err := h.FindByTrashed(context.Background(), &pb.FindAllRoleRequest{Page: 1, PageSize: 10})

	require.NoError(t, err)
	require.Len(t, res.Data, 2)
	require.Equal(t, int32(7), res.Data[0].Id)
	require.Equal(t, "2026-01-03", res.Data[0].GetDeletedAt().GetValue())
	require.Equal(t, int32(8), res.Data[1].Id)
	require.Nil(t, res.Data[1].GetDeletedAt())
}

func TestRoleQueryHandleGrpcFindByActive_NullSafeDeletedAt(t *testing.T) {
	createdAt := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.February, 2, 0, 0, 0, 0, time.UTC)

	stub := &roleQueryServiceStub{
		active: []*models.Role{
			{RoleID: 9, RoleName: "ROLE_ACTIVE", CreatedAt: &createdAt, UpdatedAt: &updatedAt},
			nil,
		},
	}

	h := NewRoleQueryHandleGrpc(stub)
	res, err := h.FindByActive(context.Background(), &pb.FindAllRoleRequest{Page: 1, PageSize: 10})

	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Equal(t, int32(9), res.Data[0].Id)
	require.Nil(t, res.Data[0].GetDeletedAt())
}
