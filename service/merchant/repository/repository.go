package repository

import (
	pb_user "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"gorm.io/gorm"
)

type Repositories interface {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository
	MerchantTransactionRepository
	UserRepository
}

type repositories struct {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository
	MerchantTransactionRepository
	UserRepository
}

func NewRepositories(db *gorm.DB, userClient pb_user.UserQueryServiceClient) Repositories {
	return &repositories{
		NewMerchantQueryRepository(db),
		NewMerchantCommandRepository(db),
		NewMerchantDocumentQueryRepository(db),
		NewMerchantDocumentCommandRepository(db),
		NewMerchantTransactionRepository(db),
		NewUserRepository(userClient),
	}
}
