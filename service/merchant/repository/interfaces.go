package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type UserRepository interface {
	FindById(ctx context.Context, user_id int) (*models.User, error)
}

type MerchantQueryRepository interface {
	FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error)
	FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error)
	FindByMerchantId(ctx context.Context, merchant_id int) (*models.Merchant, error)
	FindByName(ctx context.Context, name string) (*models.Merchant, error)
	FindByMerchantUserId(ctx context.Context, user_id int) ([]*models.Merchant, error)
	CountAllMerchants(ctx context.Context, search string) (int64, error)
	CountActiveMerchants(ctx context.Context, search string) (int64, error)
	CountTrashedMerchants(ctx context.Context, search string) (int64, error)
}

type MerchantDocumentQueryRepository interface {
	FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error)
	FindByIdDocument(ctx context.Context, id int) (*models.MerchantDocument, error)
	FindByActiveDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error)
	FindByTrashedDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error)
	CountAllDocuments(ctx context.Context, search string) (int64, error)
	CountActiveDocuments(ctx context.Context, search string) (int64, error)
	CountTrashedDocuments(ctx context.Context, search string) (int64, error)
}

type MerchantTransactionRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*models.Transaction, error)
	FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*models.Transaction, error)
	FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*models.Transaction, error)
	CountAllTransactions(ctx context.Context, search string) (int64, error)
	CountTransactionsByMerchant(ctx context.Context, merchantID int, search string) (int64, error)
	CountTransactionsByApikey(ctx context.Context, apiKey, search string) (int64, error)
}

type MerchantCommandRepository interface {
	CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error)
	UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error)
	UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error)

	TrashedMerchant(ctx context.Context, merchantId int) (*models.Merchant, error)
	RestoreMerchant(ctx context.Context, merchantId int) (*models.Merchant, error)
	DeleteMerchantPermanent(ctx context.Context, merchantId int) (bool, error)

	RestoreAllMerchant(ctx context.Context) (bool, error)
	DeleteAllMerchantPermanent(ctx context.Context) (bool, error)
}

type MerchantDocumentCommandRepository interface {
	CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error)
	TrashedMerchantDocument(ctx context.Context, merchant_document_id int) (*models.MerchantDocument, error)
	RestoreMerchantDocument(ctx context.Context, merchant_document_id int) (*models.MerchantDocument, error)
	DeleteMerchantDocumentPermanent(ctx context.Context, merchant_document_id int) (bool, error)
	RestoreAllMerchantDocument(ctx context.Context) (bool, error)
	DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error)
}
