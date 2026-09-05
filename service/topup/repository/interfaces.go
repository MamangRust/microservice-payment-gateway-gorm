package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	saldoRepository "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
)

// TopupQueryResult is the paginated topup list response.
type TopupQueryResult struct {
	Topup      *models.Topup
	TotalCount int64
}

// TopupByCardResult is the paginated topup list by card number.
type TopupByCardResult struct {
	Topup      *models.Topup
	TotalCount int64
}

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*saldoRepository.SaldoMutationResult, error)
}

type TopupQueryRepository interface {
	FindAllTopups(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error)
	FindAllTopupByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*TopupByCardResult, error)

	FindById(ctx context.Context, topup_id int) (*models.Topup, error)
}

type TopupCommandRepository interface {
	CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*models.Topup, error)
	UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*models.Topup, error)

	UpdateTopupAmount(ctx context.Context, request *requests.UpdateTopupAmount) (*models.Topup, error)
	UpdateTopupStatus(ctx context.Context, request *requests.UpdateTopupStatus) (*models.Topup, error)

	TrashedTopup(ctx context.Context, topup_id int) (*models.Topup, error)
	RestoreTopup(ctx context.Context, topup_id int) (*models.Topup, error)
	DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTopup(ctx context.Context) (bool, error)
	DeleteAllTopupPermanent(ctx context.Context) (bool, error)
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error)
}
