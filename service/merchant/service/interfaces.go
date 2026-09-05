package service

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type MerchantQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, error)
	FindById(ctx context.Context, merchant_id int) (*models.Merchant, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, error)
	FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error)
	FindByMerchantUserId(ctx context.Context, user_id int) ([]*models.Merchant, error)
}

type MerchantDocumentQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, *int, error)
	FindById(ctx context.Context, document_id int) (*models.MerchantDocument, error)
}

type MerchantTransactionService interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*models.Transaction, *int, error)
	FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*models.Transaction, *int, error)
	FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*models.Transaction, *int, error)
}

type MerchantCommandService interface {
	CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error)
	UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error)
	UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error)
	TrashedMerchant(ctx context.Context, merchant_id int) (*models.Merchant, error)
	RestoreMerchant(ctx context.Context, merchant_id int) (*models.Merchant, error)
	DeleteMerchantPermanent(ctx context.Context, merchant_id int) (bool, error)
	RestoreAllMerchant(ctx context.Context) (bool, error)
	DeleteAllMerchantPermanent(ctx context.Context) (bool, error)
}

type MerchantDocumentCommandService interface {
	CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error)
	UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error)
	TrashedMerchantDocument(ctx context.Context, document_id int) (*models.MerchantDocument, error)
	RestoreMerchantDocument(ctx context.Context, document_id int) (*models.MerchantDocument, error)
	DeleteMerchantDocumentPermanent(ctx context.Context, document_id int) (bool, error)
	RestoreAllMerchantDocument(ctx context.Context) (bool, error)
	DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error)
}
