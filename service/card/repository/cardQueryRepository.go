package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type cardQueryRepository struct {
	db *gorm.DB
}

func NewCardQueryRepository(db *gorm.DB) CardQueryRepository {
	return &cardQueryRepository{db: db}
}

func (r *cardQueryRepository) FindAllCards(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error) {
	offset := (req.Page - 1) * req.PageSize
	var cards []*models.Card
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("card_id ASC").Find(&cards).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all cards").WithInternal(err)
	}
	return cards, nil
}

func (r *cardQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error) {
	offset := (req.Page - 1) * req.PageSize
	var cards []*models.Card
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("card_id ASC").Find(&cards).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active cards").WithInternal(err)
	}
	return cards, nil
}

func (r *cardQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, error) {
	offset := (req.Page - 1) * req.PageSize
	var cards []*models.Card
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("card_id ASC").Find(&cards).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed cards").WithInternal(err)
	}
	return cards, nil
}

func (r *cardQueryRepository) FindById(ctx context.Context, cardID int) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_id = ? AND deleted_at IS NULL", cardID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindCardByUserId(ctx context.Context, userID int) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_number = ? AND deleted_at IS NULL", cardNumber).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).Where("card_number = ? AND deleted_at IS NULL", cardNumber).First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &card, nil
}

func (r *cardQueryRepository) CountAllCards(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *cardQueryRepository) CountActiveCards(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *cardQueryRepository) CountTrashedCards(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Card{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("card_number ILIKE ? OR card_type ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
