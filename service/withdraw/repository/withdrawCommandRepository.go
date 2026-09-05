package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type withdrawCommandRepository struct {
	db *gorm.DB
}

func NewWithdrawCommandRepository(db *gorm.DB) WithdrawCommandRepository {
	return &withdrawCommandRepository{db: db}
}

func (r *withdrawCommandRepository) CreateWithdraw(ctx context.Context, request *requests.CreateWithdrawRequest) (*models.Withdraw, error) {
	now := time.Now()
	withdraw := &models.Withdraw{
		WithdrawNo:     uuid.New().String(),
		CardNumber:     request.CardNumber,
		WithdrawAmount: int64(request.WithdrawAmount),
		WithdrawTime:   request.WithdrawTime,
		Status:         "pending",
		OperationID:    uuid.New().String(),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if err := r.db.WithContext(ctx).Create(withdraw).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create withdraw").WithInternal(err)
	}

	return withdraw, nil
}

func (r *withdrawCommandRepository) UpdateWithdraw(ctx context.Context, request *requests.UpdateWithdrawRequest) (*models.Withdraw, error) {
	var withdraw models.Withdraw
	if err := r.db.WithContext(ctx).First(&withdraw, int32(*request.WithdrawID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw")
	}

	now := time.Now()
	withdraw.CardNumber = request.CardNumber
	withdraw.WithdrawAmount = int64(request.WithdrawAmount)
	withdraw.WithdrawTime = request.WithdrawTime
	withdraw.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&withdraw).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw")
	}

	return &withdraw, nil
}

func (r *withdrawCommandRepository) UpdateWithdrawStatus(ctx context.Context, request *requests.UpdateWithdrawStatus) (*models.Withdraw, error) {
	var withdraw models.Withdraw
	if err := r.db.WithContext(ctx).First(&withdraw, int32(request.WithdrawID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw status")
	}

	now := time.Now()
	withdraw.Status = request.Status
	withdraw.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&withdraw).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw status")
	}

	return &withdraw, nil
}

func (r *withdrawCommandRepository) TrashedWithdraw(ctx context.Context, withdraw_id int) (*models.Withdraw, error) {
	var withdraw models.Withdraw
	if err := r.db.WithContext(ctx).First(&withdraw, int32(withdraw_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "trash withdraw")
	}

	now := time.Now()
	withdraw.DeletedAt = &now

	if err := r.db.WithContext(ctx).Save(&withdraw).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "trash withdraw")
	}

	return &withdraw, nil
}

func (r *withdrawCommandRepository) RestoreWithdraw(ctx context.Context, withdraw_id int) (*models.Withdraw, error) {
	var withdraw models.Withdraw
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").First(&withdraw, int32(withdraw_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "restore withdraw")
	}

	withdraw.DeletedAt = nil

	if err := r.db.WithContext(ctx).Unscoped().Save(&withdraw).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "restore withdraw")
	}

	return &withdraw, nil
}

func (r *withdrawCommandRepository) DeleteWithdrawPermanent(ctx context.Context, withdraw_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("withdraw_id = ?", int32(withdraw_id)).Delete(&models.Withdraw{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "withdraw", "delete withdraw permanently")
	}
	return true, nil
}

func (r *withdrawCommandRepository) RestoreAllWithdraw(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Withdraw{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all withdraws").WithInternal(result.Error)
	}
	return true, nil
}

func (r *withdrawCommandRepository) DeleteAllWithdrawPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Withdraw{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all withdraws permanently").WithInternal(result.Error)
	}
	return true, nil
}
