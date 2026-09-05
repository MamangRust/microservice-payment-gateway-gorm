package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type RoleQueryRepository interface {
	FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error)
	FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error)
	FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error)
	FindById(ctx context.Context, role_id int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByUserId(ctx context.Context, user_id int) ([]*models.Role, error)
	CountAllRoles(ctx context.Context, search string) (int64, error)
	CountActiveRoles(ctx context.Context, search string) (int64, error)
	CountTrashedRoles(ctx context.Context, search string) (int64, error)
}

type RoleCommandRepository interface {
	CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*models.Role, error)
	UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*models.Role, error)
	CreateUserRole(ctx context.Context, userID, roleID int) (*models.Role, error)
	DeleteUserRole(ctx context.Context, userID, roleID int) (bool, error)
	TrashedRole(ctx context.Context, role_id int) (*models.Role, error)
	RestoreRole(ctx context.Context, role_id int) (*models.Role, error)
	DeleteRolePermanent(ctx context.Context, role_id int) (bool, error)
	RestoreAllRole(ctx context.Context) (bool, error)
	DeleteAllRolePermanent(ctx context.Context) (bool, error)
}
