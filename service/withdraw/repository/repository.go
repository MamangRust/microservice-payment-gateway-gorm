package repository

import (
	"gorm.io/gorm"
)

type Repositories interface {
	CardRepository
	SaldoRepository
	WithdrawQueryRepository
	WithdrawCommandRepository
	IdempotencyRepository
	OutboxRepository
}

type repositories struct {
	CardRepository
	SaldoRepository
	WithdrawQueryRepository
	WithdrawCommandRepository
	IdempotencyRepository
	OutboxRepository
}

func NewRepositories(
	db *gorm.DB,
	card CardRepository,
	saldo SaldoRepository,
) Repositories {
	return &repositories{
		CardRepository:            card,
		SaldoRepository:           saldo,
		WithdrawQueryRepository:   NewWithdrawQueryRepository(db),
		WithdrawCommandRepository: NewWithdrawCommandRepository(db),
		IdempotencyRepository:     NewWithdrawIdempotencyRepository(db),
		OutboxRepository:          NewOutboxGormStore(db),
	}
}
