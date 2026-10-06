package repository

import (
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbmerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"gorm.io/gorm"
)

type GuardOptions struct {
	Saldo    []adapter.GuardOption
	Card     []adapter.GuardOption
	Merchant []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository
}

type repositories struct {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository
}

func NewRepositories(
	db *gorm.DB,
	saldoQuery pbsaldo.SaldoQueryServiceClient,
	saldoCommand pbsaldo.SaldoCommandServiceClient,
	cardQuery pbcard.CardQueryServiceClient,
	cardCommand pbcard.CardCommandServiceClient,
	merchantQuery pbmerchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoRepository:              adapter.NewSaldoAdapter(saldoQuery, saldoCommand, g.Saldo...),
		MerchantRepository:           adapter.NewMerchantAdapter(merchantQuery, g.Merchant...),
		CardRepository:               adapter.NewCardAdapter(cardQuery, cardCommand, g.Card...),
		TransactionQueryRepository:   NewTransactionQueryRepository(db),
		TransactionCommandRepository: NewTransactionCommandRepository(db),
		IdempotencyRepository:        NewTransactionIdempotencyRepository(db),
		OutboxRepository:             NewOutboxGormStore(db),
	}
}
