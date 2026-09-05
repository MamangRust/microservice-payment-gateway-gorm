package repository

import (
	"gorm.io/gorm"
)

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
	saldo SaldoRepository,
	card CardRepository,
	merchant MerchantRepository,
) Repositories {
	return &repositories{
		SaldoRepository:              saldo,
		MerchantRepository:           merchant,
		CardRepository:               card,
		TransactionQueryRepository:   NewTransactionQueryRepository(db),
		TransactionCommandRepository: NewTransactionCommandRepository(db),
		IdempotencyRepository:        NewTransactionIdempotencyRepository(db),
		OutboxRepository:             NewOutboxGormStore(db),
	}
}
