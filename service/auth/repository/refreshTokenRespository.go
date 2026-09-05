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

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND deleted_at IS NULL", token).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrTokenNotFound.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) FindByUserId(ctx context.Context, userID int) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrFindByUserID.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrParseDate.WithInternal(err)
	}

	rt := &models.RefreshToken{
		UserID:     int32(req.UserId),
		Token:      req.Token,
		Expiration: expirationTime,
	}
	err = r.db.WithContext(ctx).Create(rt).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("refresh token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create refresh token").WithInternal(err)
	}
	return rt, nil
}

func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrParseDate.WithInternal(err)
	}

	var rt models.RefreshToken
	err = r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", req.UserId).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("update refresh token").WithInternal(err)
	}

	err = r.db.WithContext(ctx).Model(&rt).Updates(map[string]interface{}{
		"token":      req.Token,
		"expiration": expirationTime,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("refresh token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update refresh token").WithInternal(err)
	}

	// Re-read
	err = r.db.WithContext(ctx).Where("refresh_token_id = ?", rt.RefreshTokenID).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("re-read refresh token").WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&models.RefreshToken{})
	if result.Error != nil {
		return sharedErrors.ErrFailed("delete refresh token").WithInternal(result.Error)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, userID int) error {
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.RefreshToken{})
	if result.Error != nil {
		return sharedErrors.ErrFailed("delete refresh token by user ID").WithInternal(result.Error)
	}
	return nil
}
