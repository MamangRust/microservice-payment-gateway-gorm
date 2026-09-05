package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type TopupQueryCache interface {
	GetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupsCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, bool)
	SetCacheTopupByCardCache(ctx context.Context, req *requests.FindAllTopupsByCardNumber, data []*repository.TopupByCardResult, total *int)

	GetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupActiveCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, bool)
	SetCachedTopupTrashedCache(ctx context.Context, req *requests.FindAllTopups, data []*repository.TopupQueryResult, total *int)

	GetCachedTopupCache(ctx context.Context, id int) (*models.Topup, bool)
	SetCachedTopupCache(ctx context.Context, data *models.Topup)
}

type TopupCommandCache interface {
	DeleteCachedTopupCache(ctx context.Context, id int)
}
