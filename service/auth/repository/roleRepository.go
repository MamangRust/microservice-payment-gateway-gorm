package repository

import (
	"context"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

type roleRepository struct {
	roleQueryClient pb.RoleQueryServiceClient
}

func NewRoleRepository(queryClient pb.RoleQueryServiceClient) *roleRepository {
	return &roleRepository{
		roleQueryClient: queryClient,
	}
}

func (r *roleRepository) FindById(ctx context.Context, id int) (*models.Role, error) {
	resp, err := r.roleQueryClient.FindByIdRole(ctx, &pb.FindByIdRoleRequest{
		RoleId: int32(id),
	})
	if err != nil {
		return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
	}

	role := resp.GetData()
	createdAt, _ := time.Parse(time.RFC3339, role.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, role.UpdatedAt)

	return &models.Role{
		RoleID:    int32(role.Id),
		RoleName:  role.Name,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	resp, err := r.roleQueryClient.FindAllRole(ctx, &pb.FindAllRoleRequest{
		Search:   name,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
	}

	roles := resp.GetData()
	if len(roles) == 0 {
		return nil, sharedErrors.ErrRoleNotFound
	}

	role := roles[0]
	createdAt, _ := time.Parse(time.RFC3339, role.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, role.UpdatedAt)

	return &models.Role{
		RoleID:    int32(role.Id),
		RoleName:  role.Name,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}
