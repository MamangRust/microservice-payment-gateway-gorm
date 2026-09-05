package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type UserQueryCache interface {
	GetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, *int, bool)
	SetCachedUsersCache(ctx context.Context, req *requests.FindAllUsers, data []*models.User, total *int)

	GetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, *int, bool)
	SetCachedUserActiveCache(ctx context.Context, req *requests.FindAllUsers, data []*models.User, total *int)

	GetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, *int, bool)
	SetCachedUserTrashedCache(ctx context.Context, req *requests.FindAllUsers, data []*models.User, total *int)

	GetCachedUserCache(ctx context.Context, id int) (*models.User, bool)
	SetCachedUserCache(ctx context.Context, data *models.User)
}

type UserCommandCache interface {
	DeleteUserCache(ctx context.Context, id int)
}
