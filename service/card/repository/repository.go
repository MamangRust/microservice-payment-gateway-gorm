package repository

import (
	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"gorm.io/gorm"
)

type GuardOptions struct {
	User []adapter.GuardOption
}

type Repositories struct {
	CardCommand         CardCommandRepository
	CardQuery           CardQueryRepository
	User                adapter.UserAdapter
	CardAuthTransaction CardAuthTransactionRepository
	CardPayment         CardPaymentRepository
	CardReward          CardRewardRepository
	BillingCycle        BillingCycleRepository
}

func NewRepositories(db *gorm.DB, userQueryClient pbuser.UserQueryServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		CardQuery:           NewCardQueryRepository(db),
		CardCommand:         NewCardCommandRepository(db),
		User:                adapter.NewUserAdapter(userQueryClient, g.User...),
		CardAuthTransaction: NewCardAuthTransactionRepository(db),
		CardPayment:         NewCardPaymentRepository(db),
		CardReward:          NewCardRewardRepository(db),
		BillingCycle:        NewBillingCycleRepository(db),
	}
}
