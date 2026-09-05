package repository

import (
	"context"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

type userRoleRepository struct {
	roleCommandClient pb.RoleCommandServiceClient
}

func NewUserRoleRepository(commandClient pb.RoleCommandServiceClient) *userRoleRepository {
	return &userRoleRepository{
		roleCommandClient: commandClient,
	}
}

func (r *userRoleRepository) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	_, err := r.roleCommandClient.CreateUserRole(ctx, &pb.CreateUserRoleRequest{
		UserId: int32(req.UserId),
		RoleId: int32(req.RoleId),
	})
	if err != nil {
		return nil, sharedErrors.ErrFailed("assign role to user").WithInternal(err)
	}

	return &models.UserRole{
		UserID: int32(req.UserId),
		RoleID: int32(req.RoleId),
	}, nil
}

func (r *userRoleRepository) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	_, err := r.roleCommandClient.DeleteUserRole(ctx, &pb.DeleteUserRoleRequest{
		UserId: int32(req.UserId),
		RoleId: int32(req.RoleId),
	})
	if err != nil {
		return sharedErrors.ErrFailed("remove role from user").WithInternal(err)
	}
	return nil
}
