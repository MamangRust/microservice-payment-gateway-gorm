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

type transactionCommandRepository struct {
	db *gorm.DB
}

func NewTransactionCommandRepository(db *gorm.DB) TransactionCommandRepository {
	return &transactionCommandRepository{db: db}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*models.Transaction, error) {
	now := time.Now()
	txn := &models.Transaction{
		TransactionNo:   uuid.New().String(),
		CardNumber:      request.CardNumber,
		Amount:          int64(request.Amount),
		PaymentMethod:   request.PaymentMethod,
		MerchantID:      int32(*request.MerchantID),
		TransactionTime: request.TransactionTime,
		Status:          "pending",
		OperationID:     uuid.New().String(),
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}

	if err := r.db.WithContext(ctx).Create(txn).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create transaction").WithInternal(err)
	}

	return txn, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*models.Transaction, error) {
	var txn models.Transaction
	if err := r.db.WithContext(ctx).First(&txn, int32(*request.TransactionID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction")
	}

	now := time.Now()
	txn.CardNumber = request.CardNumber
	txn.Amount = int64(request.Amount)
	txn.PaymentMethod = request.PaymentMethod
	txn.MerchantID = int32(*request.MerchantID)
	txn.TransactionTime = request.TransactionTime
	txn.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction")
	}

	return &txn, nil
}

func (r *transactionCommandRepository) UpdateTransactionStatus(ctx context.Context, request *requests.UpdateTransactionStatus) (*models.Transaction, error) {
	var txn models.Transaction
	if err := r.db.WithContext(ctx).First(&txn, int32(request.TransactionID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction status")
	}

	now := time.Now()
	txn.Status = request.Status
	txn.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction status")
	}

	return &txn, nil
}

func (r *transactionCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}

	result := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("transaction_id = ? AND status = ?", int32(id), fromStatus).
		Update("status", toStatus)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (r *transactionCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*models.Transaction, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}

	var txn models.Transaction
	err := r.db.WithContext(ctx).
		Where("transaction_id = ? AND status = ?", int32(id), fromStatus).
		First(&txn).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	txn.Status = toStatus
	txn.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, err
	}

	return &txn, nil
}

func (r *transactionCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*models.Transaction, error) {
	threshold := time.Now().Add(-olderThan)
	var txns []*models.Transaction
	err := r.db.WithContext(ctx).
		Where("status IN ('processing', 'compensating') AND updated_at < ?", threshold).
		Limit(int(maxRows)).
		Find(&txns).Error
	if err != nil {
		return nil, err
	}
	return txns, nil
}

func (r *transactionCommandRepository) TrashedTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var txn models.Transaction
	if err := r.db.WithContext(ctx).First(&txn, int32(transaction_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "trash transaction")
	}

	now := time.Now()
	txn.DeletedAt = &now

	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "trash transaction")
	}

	return &txn, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var txn models.Transaction
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").First(&txn, int32(transaction_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "restore transaction")
	}

	txn.DeletedAt = nil

	if err := r.db.WithContext(ctx).Unscoped().Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "restore transaction")
	}

	return &txn, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("transaction_id = ?", int32(transaction_id)).Delete(&models.Transaction{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "transaction", "delete transaction permanently")
	}
	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransaction(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Transaction{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all transactions").WithInternal(result.Error)
	}
	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Transaction{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all transactions permanently").WithInternal(result.Error)
	}
	return true, nil
}
