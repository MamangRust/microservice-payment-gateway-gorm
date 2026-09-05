package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type billingCycleRepository struct {
	db *gorm.DB
}

func NewBillingCycleRepository(db *gorm.DB) BillingCycleRepository {
	return &billingCycleRepository{db: db}
}

func (r *billingCycleRepository) GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*models.BillingCycle, error) {
	var cycles []*models.BillingCycle
	err := r.db.WithContext(ctx).
		Where("card_number = ?", cardNumber).
		Order("cycle_start DESC").Find(&cycles).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get billing cycles").WithInternal(err)
	}
	return cycles, nil
}

func (r *billingCycleRepository) CreateBillingCycles(ctx context.Context, cycleStart, cycleEnd, dueDate time.Time) (int, error) {
	var activeCards []models.Card
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&activeCards).Error; err != nil {
		return 0, sharedErrors.ErrFailed("find active cards for billing cycles").WithInternal(err)
	}

	var created int
	for _, card := range activeCards {
		cycle := models.BillingCycle{
			CardNumber: card.CardNumber,
			CycleStart: cycleStart,
			CycleEnd:   cycleEnd,
			DueDate:    dueDate,
			Status:     "open",
		}
		if err := r.db.WithContext(ctx).Create(&cycle).Error; err != nil {
			if database.IsUniqueViolation(err) {
				continue
			}
			return created, sharedErrors.ErrFailed("create billing cycle").WithInternal(err)
		}
		created++
	}
	return created, nil
}
