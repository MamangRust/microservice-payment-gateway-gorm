package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cardPaymentRepository struct {
	db *gorm.DB
}

func NewCardPaymentRepository(db *gorm.DB) CardPaymentRepository {
	return &cardPaymentRepository{db: db}
}

func (r *cardPaymentRepository) PostPayment(ctx context.Context, req *requests.PostPaymentRequest) (*models.CardPayment, error) {
	var billingID *int32
	if req.BillingID != nil {
		billingID = new(int32)
		*billingID = int32(*req.BillingID)
	}

	payment := &models.CardPayment{
		PaymentUuid:    uuid.New().String(),
		CardNumber:     req.CardNumber,
		BillingID:      billingID,
		Amount:         req.Amount,
		PaymentChannel: req.PaymentChannel,
		ReferenceID:    &req.ReferenceID,
	}

	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create card payment").WithInternal(err)
	}
	return payment, nil
}

func (r *cardPaymentRepository) GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*models.CardPayment, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	var payments []*models.CardPayment
	err := r.db.WithContext(ctx).Where("card_number = ?", cardNumber).
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&payments).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get card payments").WithInternal(err)
	}
	return payments, nil
}

func (r *cardPaymentRepository) CountPayments(ctx context.Context, cardNumber string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CardPayment{}).
		Where("card_number = ?", cardNumber).Count(&count).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("count card payments").WithInternal(err)
	}
	return int(count), nil
}
