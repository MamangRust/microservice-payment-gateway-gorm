package repository

import (
	"gorm.io/gorm"
)

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
	saldo SaldoRepository,
	card CardRepository,
) Repositories {
	return &repositories{
		SaldoRepository:           saldo,
		TransferQueryRepository:   NewTransferQueryRepository(db),
		TransferCommandRepository: NewTransferCommandRepository(db),
		CardRepository:            card,
		IdempotencyRepository:     NewTransferIdempotencyRepository(db),
		OutboxRepository:          NewOutboxGormStore(db),
	}
}
