package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type cardAuthTransactionRepository struct {
	db *gorm.DB
}

func NewCardAuthTransactionRepository(db *gorm.DB) CardAuthTransactionRepository {
	return &cardAuthTransactionRepository{db: db}
}

func (r *cardAuthTransactionRepository) InsertPending(ctx context.Context, req *requests.AuthorizeCardRequest) (*models.CardAuthTransaction, error) {
	txn := &models.CardAuthTransaction{
		TxnID:          req.IdempotencyKey + "-txn",
		CardNumber:     req.CardNumber,
		MerchantID:     int32(req.MerchantID),
		Amount:         req.Amount,
		Currency:       req.Currency,
		Mcc:            req.Mcc,
		PosEntryMode:   req.PosEntryMode,
		IdempotencyKey: req.IdempotencyKey,
		Status:         "pending",
	}
	if err := r.db.WithContext(ctx).Create(txn).Error; err != nil {
		return nil, sharedErrors.ErrFailed("insert authorization transaction").WithInternal(err)
	}
	return txn, nil
}

func (r *cardAuthTransactionRepository) Approve(ctx context.Context, txnID string) (*models.CardAuthTransaction, error) {
	var txn models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("txn_id = ?", txnID).First(&txn).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("approve authorization transaction").WithInternal(err)
	}
	txn.Status = "approved"
	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrFailed("approve authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) Decline(ctx context.Context, txnID string) (*models.CardAuthTransaction, error) {
	var txn models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("txn_id = ?", txnID).First(&txn).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("decline authorization transaction").WithInternal(err)
	}
	txn.Status = "declined"
	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrFailed("decline authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) Reverse(ctx context.Context, txnID string) (*models.CardAuthTransaction, error) {
	var txn models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("txn_id = ?", txnID).First(&txn).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("reverse authorization transaction").WithInternal(err)
	}
	txn.Status = "reversed"
	if err := r.db.WithContext(ctx).Save(&txn).Error; err != nil {
		return nil, sharedErrors.ErrFailed("reverse authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (*models.CardAuthTransaction, error) {
	var txn models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&txn).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) FindByTxnID(ctx context.Context, txnID string) (*models.CardAuthTransaction, error) {
	var txn models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("txn_id = ?", txnID).First(&txn).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &txn, nil
}

func (r *cardAuthTransactionRepository) FindByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*models.CardAuthTransaction, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	var txns []*models.CardAuthTransaction
	err := r.db.WithContext(ctx).Where("card_number = ?", cardNumber).
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&txns).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transactions").WithInternal(err)
	}
	return txns, nil
}

func (r *cardAuthTransactionRepository) CountRecentByCardNumber(ctx context.Context, cardNumber string, since time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CardAuthTransaction{}).
		Where("card_number = ? AND created_at >= ?", cardNumber, since).Count(&count).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("count recent authorization transactions").WithInternal(err)
	}
	return int(count), nil
}

func (r *cardAuthTransactionRepository) UpdateRiskScore(ctx context.Context, txnID string, score int) error {
	err := r.db.WithContext(ctx).Model(&models.CardAuthTransaction{}).
		Where("txn_id = ?", txnID).Update("risk_score", score).Error
	if err != nil {
		return sharedErrors.ErrFailed("update authorization risk score").WithInternal(err)
	}
	return nil
}
