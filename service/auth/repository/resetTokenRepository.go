package repository

import (
	"context"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type resetTokenRepository struct {
	db *gorm.DB
}

func NewResetTokenRepository(db *gorm.DB) *resetTokenRepository {
	return &resetTokenRepository{db: db}
}

func isUniqueViolationRT(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*models.ResetToken, error) {
	var rt models.ResetToken
	err := r.db.WithContext(ctx).Where("token = ?", code).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*models.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt := &models.ResetToken{
		UserID:     int32(req.UserID),
		Token:      req.ResetToken,
		ExpiryDate: expiryDate,
	}
	err = r.db.WithContext(ctx).Create(rt).Error
	if err != nil {
		if isUniqueViolationRT(err) {
			return nil, sharedErrors.NewConflictError("reset token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create reset token").WithInternal(err)
	}
	return rt, nil
}

func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, userID int32) error {
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.ResetToken{})
	if result.Error != nil {
		return sharedErrors.ErrFailed("delete reset token").WithInternal(result.Error)
	}
	return nil
}
