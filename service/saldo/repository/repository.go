package repository

import (
	"gorm.io/gorm"
)

type Repositories interface {
	SaldoQueryRepository
	SaldoCommandRepository
	CardRepository
}

type repositories struct {
	SaldoQueryRepository
	SaldoCommandRepository
	CardRepository
}

func NewRepositories(db *gorm.DB, card CardRepository) Repositories {
	return &repositories{
		SaldoQueryRepository:   NewSaldoQueryRepository(db),
		SaldoCommandRepository: NewSaldoCommandRepository(db),
		CardRepository:         card,
	}
}
