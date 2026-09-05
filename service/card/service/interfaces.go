package service

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type CardQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, error)
	FindById(ctx context.Context, cardID int) (*models.Card, error)
	FindByUserID(ctx context.Context, userID int) (*models.Card, error)
	FindByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error)
	FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error)
}

type CardCommandService interface {
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
	ProcessBillingCycles(ctx context.Context) error
}

type BillingEngineService interface {
	TriggerBillingCycle(ctx context.Context, billingCycleDay int) (int, error)
	GetStatement(ctx context.Context, cardNumber string) (*models.BillingCycle, error)
	GetStatementsByCard(ctx context.Context, cardNumber string, page, pageSize int) ([]*models.BillingCycle, error)
	GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*models.BillingCycle, error)
}
