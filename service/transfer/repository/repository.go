package repository

import (
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"gorm.io/gorm"
)

type GuardOptions struct {
	Saldo []adapter.GuardOption
	Card  []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	IdempotencyRepository
	OutboxRepository
}

type repositories struct {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	IdempotencyRepository
	OutboxRepository
}

func NewRepositories(
	db *gorm.DB,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCommand pbsaldo.SaldoCommandServiceClient,
	cardQuery pbcard.CardQueryServiceClient,
	cardCommand pbcard.CardCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}
	return &repositories{
		SaldoRepository:           adapter.NewSaldoAdapter(saldoQuery, saldoCommand, g.Saldo...),
		TransferQueryRepository:   NewTransferQueryRepository(db),
		TransferCommandRepository: NewTransferCommandRepository(db),
		CardRepository:            adapter.NewCardAdapter(cardQuery, cardCommand, g.Card...),
		IdempotencyRepository:     NewTransferIdempotencyRepository(db),
		OutboxRepository:          NewOutboxGormStore(db),
	}
}
