package repository

import (
	"gorm.io/gorm"
)

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
	card CardRepository,
	saldo SaldoRepository,
) Repositories {
	return &repositories{
		TopupQueryRepository:   NewTopupQueryRepository(db),
		TopupCommandRepository: NewTopupCommandRepository(db),
		CardRepository:         card,
		SaldoRepository:        saldo,
		IdempotencyRepository:  NewTopupIdempotencyRepository(db),
		OutboxRepository:       NewOutboxGormStore(db),
	}
}
