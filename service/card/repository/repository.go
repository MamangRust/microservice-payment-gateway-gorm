package repository

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"gorm.io/gorm"
)

// Repositories contains all repositories used by the card service.
type Repositories struct {
	CardCommand         CardCommandRepository
	CardQuery           CardQueryRepository
	User                UserRepository
	CardAuthTransaction CardAuthTransactionRepository
	CardPayment         CardPaymentRepository
	CardReward          CardRewardRepository
	BillingCycle        BillingCycleRepository
}

func NewRepositories(db *gorm.DB, userClient pb.UserQueryServiceClient) *Repositories {
	return &Repositories{
		CardQuery:           NewCardQueryRepository(db),
		CardCommand:         NewCardCommandRepository(db),
		User:                NewUserRepository(userClient),
		CardAuthTransaction: NewCardAuthTransactionRepository(db),
		CardPayment:         NewCardPaymentRepository(db),
		CardReward:          NewCardRewardRepository(db),
		BillingCycle:        NewBillingCycleRepository(db),
	}
}
