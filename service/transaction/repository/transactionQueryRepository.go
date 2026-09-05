package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type transactionQueryRepository struct {
	db *gorm.DB
}

func NewTransactionQueryRepository(db *gorm.DB) TransactionQueryRepository {
	return &transactionQueryRepository{db: db}
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{})
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count transactions").WithInternal(err)
	}

	var txns []*models.Transaction
	query2 := r.db.WithContext(ctx)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transaction_time DESC").Offset(offset).Limit(req.PageSize).Find(&txns).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all transactions").WithInternal(err)
	}

	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count active transactions").WithInternal(err)
	}

	var txns []*models.Transaction
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transaction_time DESC").Offset(offset).Limit(req.PageSize).Find(&txns).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active transactions").WithInternal(err)
	}

	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count trashed transactions").WithInternal(err)
	}

	var txns []*models.Transaction
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR payment_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transaction_time DESC").Offset(offset).Limit(req.PageSize).Find(&txns).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed transactions").WithInternal(err)
	}

	results := make([]*TransactionQueryResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionQueryResult{Transaction: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindAllTransactionByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*TransactionByCardResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("payment_method LIKE ? OR status LIKE ?", like, like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count transactions by card number").WithInternal(err)
	}

	var txns []*models.Transaction
	query2 := r.db.WithContext(ctx).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("payment_method LIKE ? OR status LIKE ?", like, like)
	}
	if err := query2.Order("transaction_time DESC").Offset(offset).Limit(req.PageSize).Find(&txns).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find transactions by card number").WithInternal(err)
	}

	results := make([]*TransactionByCardResult, len(txns))
	for i, t := range txns {
		results[i] = &TransactionByCardResult{Transaction: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var txn models.Transaction
	err := r.db.WithContext(ctx).First(&txn, int32(transaction_id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &txn, nil
}

func (r *transactionQueryRepository) FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*models.Transaction, error) {
	var txns []*models.Transaction
	err := r.db.WithContext(ctx).Where("merchant_id = ?", int32(merchant_id)).Order("transaction_time DESC").Find(&txns).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return txns, nil
}
