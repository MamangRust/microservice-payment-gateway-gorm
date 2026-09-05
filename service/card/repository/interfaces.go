package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type CardCommandRepository interface {
	CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*models.Card, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error)
	TrashedCard(ctx context.Context, cardID int) (*models.Card, error)
	RestoreCard(ctx context.Context, cardID int) (*models.Card, error)
	DeleteCardPermanent(ctx context.Context, cardID int) (bool, error)
	RestoreAllCard(ctx context.Context) (bool, error)
	DeleteAllCardPermanent(ctx context.Context) (bool, error)
	ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*models.Card, error)
	UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*models.Card, error)
	RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*models.Card, error)
}

type CardQueryRepository interface {
	FindAllCards(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error)
	FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error)
	FindById(ctx context.Context, cardID int) (*models.Card, error)
	FindCardByUserId(ctx context.Context, userID int) (*models.Card, error)
	FindCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error)
	FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error)
	CountAllCards(ctx context.Context, search string) (int64, error)
	CountActiveCards(ctx context.Context, search string) (int64, error)
	CountTrashedCards(ctx context.Context, search string) (int64, error)
}

type UserRepository interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
}

type CardAuthTransactionRepository interface {
	InsertPending(ctx context.Context, req *requests.AuthorizeCardRequest) (*models.CardAuthTransaction, error)
	Approve(ctx context.Context, txnID string) (*models.CardAuthTransaction, error)
	Decline(ctx context.Context, txnID string) (*models.CardAuthTransaction, error)
	Reverse(ctx context.Context, txnID string) (*models.CardAuthTransaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*models.CardAuthTransaction, error)
	FindByTxnID(ctx context.Context, txnID string) (*models.CardAuthTransaction, error)
	FindByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*models.CardAuthTransaction, error)
	CountRecentByCardNumber(ctx context.Context, cardNumber string, since time.Time) (int, error)
	UpdateRiskScore(ctx context.Context, txnID string, score int) error
}

type CardPaymentRepository interface {
	PostPayment(ctx context.Context, req *requests.PostPaymentRequest) (*models.CardPayment, error)
	GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*models.CardPayment, error)
	CountPayments(ctx context.Context, cardNumber string) (int, error)
}

type CardRewardRepository interface {
	EarnRewards(ctx context.Context, req *requests.EarnRewardsRequest) (*models.CardReward, error)
	GetBalance(ctx context.Context, cardNumber string) (int64, error)
	GetHistory(ctx context.Context, cardNumber string) ([]*models.CardReward, error)
	RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error)
}

type BillingCycleRepository interface {
	GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*models.BillingCycle, error)
	CreateBillingCycles(ctx context.Context, cycleStart, cycleEnd, dueDate time.Time) (int, error)
}
