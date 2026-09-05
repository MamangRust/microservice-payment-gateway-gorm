package repository

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type merchantTransactionRepository struct {
	db *gorm.DB
}

func NewMerchantTransactionRepository(db *gorm.DB) MerchantTransactionRepository {
	return &merchantTransactionRepository{db: db}
}

func (r *merchantTransactionRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*models.Transaction, error) {
	offset := (req.Page - 1) * req.PageSize
	var transactions []*models.Transaction
	query := r.db.WithContext(ctx).Model(&models.Transaction{})
	if req.Search != "" {
		query = query.Where("card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Joins("LEFT JOIN merchants ON merchants.merchant_id = transactions.merchant_id").
		Select("transactions.*, COALESCE(merchants.name, '') AS merchant_name").
		Offset(offset).Limit(req.PageSize).Order("transactions.transaction_id ASC").Find(&transactions).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all merchant transactions").WithInternal(err)
	}
	return transactions, nil
}

func (r *merchantTransactionRepository) FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*models.Transaction, error) {
	offset := (req.Page - 1) * req.PageSize
	var transactions []*models.Transaction
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("transactions.merchant_id = ?", req.MerchantID)
	if req.Search != "" {
		query = query.Where("transactions.card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Joins("LEFT JOIN merchants ON merchants.merchant_id = transactions.merchant_id").
		Select("transactions.*, COALESCE(merchants.name, '') AS merchant_name").
		Offset(offset).Limit(req.PageSize).Order("transactions.transaction_id ASC").Find(&transactions).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find merchant transactions by merchant").WithInternal(err)
	}
	return transactions, nil
}

func (r *merchantTransactionRepository) FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*models.Transaction, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("api_key = ?", req.ApiKey).First(&merchant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*models.Transaction{}, nil
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var transactions []*models.Transaction
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("transactions.merchant_id = ?", merchant.MerchantID)
	if req.Search != "" {
		query = query.Where("transactions.card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Joins("LEFT JOIN merchants ON merchants.merchant_id = transactions.merchant_id").
		Select("transactions.*, COALESCE(merchants.name, '') AS merchant_name").
		Offset(offset).Limit(req.PageSize).Order("transactions.transaction_id ASC").Find(&transactions).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find merchant transactions by API key").WithInternal(err)
	}
	return transactions, nil
}

func (r *merchantTransactionRepository) CountAllTransactions(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{})
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *merchantTransactionRepository) CountTransactionsByMerchant(ctx context.Context, merchantID int, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("merchant_id = ?", merchantID)
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *merchantTransactionRepository) CountTransactionsByApikey(ctx context.Context, apiKey, search string) (int64, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&merchant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("find merchant by api key: %w", err)
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("merchant_id = ?", merchant.MerchantID)
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
