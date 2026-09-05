package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type RoleQueryCache interface {
	SetCachedRoles(ctx context.Context, req *requests.FindAllRoles, data []*models.Role, total *int)
	GetCachedRoles(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, *int, bool)

	GetCachedRoleById(ctx context.Context, id int) (*models.Role, bool)
	SetCachedRoleById(ctx context.Context, id int, data *models.Role)

	GetCachedRoleByUserId(ctx context.Context, userId int) ([]*models.Role, bool)
	SetCachedRoleByUserId(ctx context.Context, userId int, data []*models.Role)

	GetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, *int, bool)
	SetCachedRoleActive(ctx context.Context, req *requests.FindAllRoles, data []*models.Role, total *int)

	GetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, *int, bool)
	SetCachedRoleTrashed(ctx context.Context, req *requests.FindAllRoles, data []*models.Role, total *int)

	GetCachedRoleByName(ctx context.Context, name string) (*models.Role, bool)
	SetCachedRoleByName(ctx context.Context, name string, data *models.Role)
}

type RoleCommandCache interface {
	DeleteCachedRole(ctx context.Context, id int)
}
