package repository

import (
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"gorm.io/gorm"
)

type GuardOptions struct {
	Card  []adapter.GuardOption
	Saldo []adapter.GuardOption
}

type Repositories interface {
	TopupQueryRepository
	TopupCommandRepository
	CardRepository
	SaldoRepository
	IdempotencyRepository
	OutboxRepository
}

type repositories struct {
	TopupQueryRepository
	TopupCommandRepository
	CardRepository
	SaldoRepository
	IdempotencyRepository
	OutboxRepository
}

func NewRepositories(
	db *gorm.DB,
	cardQuery pbcard.CardQueryServiceClient,
	cardCommand pbcard.CardCommandServiceClient,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCommand pbsaldo.SaldoCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		TopupQueryRepository:   NewTopupQueryRepository(db),
		TopupCommandRepository: NewTopupCommandRepository(db),
		CardRepository:         adapter.NewCardAdapter(cardQuery, cardCommand, g.Card...),
		SaldoRepository:        adapter.NewSaldoAdapter(saldoQuery, saldoCommand, g.Saldo...),
		IdempotencyRepository:  NewTopupIdempotencyRepository(db),
		OutboxRepository:       NewOutboxGormStore(db),
	}
}
