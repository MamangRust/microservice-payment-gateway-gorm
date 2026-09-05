package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type CardQueryCache interface {
	GetByIdCache(ctx context.Context, cardID int) (*models.Card, bool)
	GetByUserIDCache(ctx context.Context, userID int) (*models.Card, bool)
	GetByCardNumberCache(ctx context.Context, cardNumber string) (*models.Card, bool)
	GetUserCardByCardNumberCache(ctx context.Context, cardNumber string) (*models.Card, bool)
	GetFindAllCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool)
	GetByActiveCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool)
	GetByTrashedCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool)
	SetByIdCache(ctx context.Context, cardID int, data *models.Card)
	SetByUserIDCache(ctx context.Context, userID int, data *models.Card)
	SetByCardNumberCache(ctx context.Context, cardNumber string, data *models.Card)
	SetFindAllCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, totalRecords *int)
	SetByActiveCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, totalRecords *int)
	SetUserCardByCardNumberCache(ctx context.Context, cardNumber string, data *models.Card)
	SetByTrashedCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, totalRecords *int)
	DeleteByIdCache(ctx context.Context, cardID int)
	DeleteByUserIDCache(ctx context.Context, userID int)
	DeleteByCardNumberCache(ctx context.Context, cardNumber string)
}

type CardCommandCache interface {
	DeleteCardCommandCache(ctx context.Context, id int)
}
