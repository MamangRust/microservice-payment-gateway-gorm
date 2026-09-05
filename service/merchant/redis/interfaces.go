package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type MerchantQueryCache interface {
	GetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool)
	SetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int)

	GetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool)
	SetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int)

	GetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool)
	SetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int)

	GetCachedMerchant(ctx context.Context, id int) (*models.Merchant, bool)
	SetCachedMerchant(ctx context.Context, data *models.Merchant)

	GetCachedMerchantsByUserId(ctx context.Context, userId int) ([]*models.Merchant, bool)
	SetCachedMerchantsByUserId(ctx context.Context, userId int, data []*models.Merchant)

	GetCachedMerchantByApiKey(ctx context.Context, apiKey string) (*models.Merchant, bool)
	SetCachedMerchantByApiKey(ctx context.Context, apiKey string, data *models.Merchant)
}

type MerchantDocumentQueryCache interface {
	GetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, bool)
	SetCachedMerchantDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*models.MerchantDocument, total *int)

	GetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, bool)
	SetCachedMerchantDocumentsActive(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*models.MerchantDocument, total *int)

	GetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, bool)
	SetCachedMerchantDocumentsTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments, data []*models.MerchantDocument, total *int)

	GetCachedMerchantDocument(ctx context.Context, id int) (*models.MerchantDocument, bool)
	SetCachedMerchantDocument(ctx context.Context, id int, data *models.MerchantDocument)
}

type MerchantCommandCache interface {
	DeleteCachedMerchant(ctx context.Context, id int)
}

type MerchantDocumentCommandCache interface {
	DeleteCachedMerchantDocuments(ctx context.Context, id int)
}

type MerchantTransactionCache interface {
	GetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*models.Transaction, *int, bool)
	SetCacheAllMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions, data []*models.Transaction, total *int)

	GetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*models.Transaction, *int, bool)
	SetCacheMerchantTransactions(ctx context.Context, req *requests.FindAllMerchantTransactionsById, data []*models.Transaction, total *int)

	GetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*models.Transaction, *int, bool)
	SetCacheMerchantTransactionApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey, data []*models.Transaction, total *int)
}
