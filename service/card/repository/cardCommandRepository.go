package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/randomvcc"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type cardCommandRepository struct {
	db *gorm.DB
}

func NewCardCommandRepository(db *gorm.DB) CardCommandRepository {
	return &cardCommandRepository{db: db}
}

func (r *cardCommandRepository) CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*models.Card, error) {
	number, err := randomvcc.RandomCardNumber()
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	card := &models.Card{
		UserID:       int32(request.UserID),
		CardNumber:   number,
		CardType:     request.CardType,
		ExpireDate:   request.ExpireDate,
		Cvv:          request.CVV,
		CardProvider: request.CardProvider,
		Status:       "active",
		CreditLimit:  50000000, // Default 50M
	}

	if err := r.db.WithContext(ctx).Create(card).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("card number already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create card").WithInternal(err)
	}
	return card, nil
}

func (r *cardCommandRepository) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_id = ? AND deleted_at IS NULL", request.CardID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update card").WithInternal(err)
	}

	card.CardType = request.CardType
	card.ExpireDate = request.ExpireDate
	card.Cvv = request.CVV
	card.CardProvider = request.CardProvider

	if err := r.db.WithContext(ctx).Save(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) TrashedCard(ctx context.Context, cardID int) (*models.Card, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Card{}).
		Where("card_id = ? AND deleted_at IS NULL", cardID).
		Update("deleted_at", now)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("trash card").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("card")
	}

	var card models.Card
	if err := r.db.WithContext(ctx).Where("card_id = ?", cardID).First(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get trashed card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) RestoreCard(ctx context.Context, cardID int) (*models.Card, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Card{}).
		Where("card_id = ? AND deleted_at IS NOT NULL", cardID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("restore card").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("card")
	}

	var card models.Card
	if err := r.db.WithContext(ctx).Where("card_id = ?", cardID).First(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get restored card").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) DeleteCardPermanent(ctx context.Context, cardID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("card_id = ?", cardID).Delete(&models.Card{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete card permanently").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("card")
	}
	return true, nil
}

func (r *cardCommandRepository) RestoreAllCard(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Card{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, sharedErrors.ErrFailed("restore all cards").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) DeleteAllCardPermanent(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Card{}).Error; err != nil {
		return false, sharedErrors.ErrFailed("delete all cards permanently").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_id = ? AND deleted_at IS NULL", request.CardID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("toggle card status").WithInternal(err)
	}

	if card.Status == "active" {
		card.Status = "inactive"
	} else {
		card.Status = "active"
	}

	if err := r.db.WithContext(ctx).Save(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("toggle card status").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_id = ? AND deleted_at IS NULL", request.CardID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update credit limit").WithInternal(err)
	}

	card.CreditLimit = int32(request.CreditLimit)

	if err := r.db.WithContext(ctx).Save(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update credit limit").WithInternal(err)
	}
	return &card, nil
}

func (r *cardCommandRepository) RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_id = ? AND deleted_at IS NULL", request.CardID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(err)
	}

	if card.RewardPoints < int32(request.Points) {
		return nil, sharedErrors.NewBadRequestError("insufficient reward points")
	}

	card.RewardPoints -= int32(request.Points)

	if err := r.db.WithContext(ctx).Save(&card).Error; err != nil {
		return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(err)
	}
	return &card, nil
}
