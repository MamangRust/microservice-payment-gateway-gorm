package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/state"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transferCommandRepository struct {
	db *gorm.DB
}

func NewTransferCommandRepository(db *gorm.DB) TransferCommandRepository {
	return &transferCommandRepository{db: db}
}

func (r *transferCommandRepository) CreateTransfer(ctx context.Context, request *requests.CreateTransferRequest) (*models.Transfer, error) {
	now := time.Now()
	transfer := &models.Transfer{
		TransferNo:     uuid.New().String(),
		TransferFrom:   request.TransferFrom,
		TransferTo:     request.TransferTo,
		TransferAmount: int64(request.TransferAmount),
		TransferTime:   now,
		Status:         state.Pending,
		OperationID:    uuid.New().String(),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if err := r.db.WithContext(ctx).Create(transfer).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create transfer").WithInternal(err)
	}

	return transfer, nil
}

func (r *transferCommandRepository) UpdateTransfer(ctx context.Context, request *requests.UpdateTransferRequest) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).First(&transfer, int32(*request.TransferID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer")
	}

	now := time.Now()
	transfer.TransferFrom = request.TransferFrom
	transfer.TransferTo = request.TransferTo
	transfer.TransferAmount = int64(request.TransferAmount)
	transfer.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&transfer).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer")
	}

	return &transfer, nil
}

func (r *transferCommandRepository) UpdateTransferAmount(ctx context.Context, request *requests.UpdateTransferAmountRequest) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).First(&transfer, int32(request.TransferID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer amount")
	}

	now := time.Now()
	transfer.TransferAmount = int64(request.TransferAmount)
	transfer.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&transfer).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer amount")
	}

	return &transfer, nil
}

func (r *transferCommandRepository) UpdateTransferStatus(ctx context.Context, request *requests.UpdateTransferStatus) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).First(&transfer, int32(request.TransferID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer status")
	}

	now := time.Now()
	transfer.Status = request.Status
	transfer.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&transfer).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer status")
	}

	return &transfer, nil
}

func (r *transferCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}

	result := r.db.WithContext(ctx).
		Model(&models.Transfer{}).
		Where("transfer_id = ? AND status = ? AND deleted_at IS NULL", int32(id), fromStatus).
		Update("status", toStatus)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (r *transferCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*models.Transfer, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}

	var transfer models.Transfer
	err := r.db.WithContext(ctx).
		Where("transfer_id = ? AND status = ? AND deleted_at IS NULL", int32(id), fromStatus).
		First(&transfer).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	transfer.Status = toStatus
	transfer.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&transfer).Error; err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (r *transferCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*models.Transfer, error) {
	threshold := time.Now().Add(-olderThan)
	var transfers []*models.Transfer
	err := r.db.WithContext(ctx).
		Where("status IN ('processing', 'compensating') AND updated_at < ? AND deleted_at IS NULL", threshold).
		Limit(int(maxRows)).Find(&transfers).Error
	if err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *transferCommandRepository) TrashedTransfer(ctx context.Context, transfer_id int) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).First(&transfer, int32(transfer_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "trash transfer")
	}

	now := time.Now()
	transfer.DeletedAt = &now

	if err := r.db.WithContext(ctx).Save(&transfer).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "trash transfer")
	}

	return &transfer, nil
}

func (r *transferCommandRepository) RestoreTransfer(ctx context.Context, transfer_id int) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").First(&transfer, int32(transfer_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "restore transfer")
	}

	transfer.DeletedAt = nil

	if err := r.db.WithContext(ctx).Unscoped().Save(&transfer).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "restore transfer")
	}

	return &transfer, nil
}

func (r *transferCommandRepository) DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("transfer_id = ?", int32(transfer_id)).Delete(&models.Transfer{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "transfer", "delete transfer permanently")
	}
	return true, nil
}

func (r *transferCommandRepository) RestoreAllTransfer(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Transfer{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all transfers").WithInternal(result.Error)
	}
	return true, nil
}

func (r *transferCommandRepository) DeleteAllTransferPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Transfer{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all transfers permanently").WithInternal(result.Error)
	}
	return true, nil
}
