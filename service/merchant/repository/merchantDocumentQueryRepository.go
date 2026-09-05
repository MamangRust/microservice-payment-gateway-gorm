package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type merchantDocumentQueryRepository struct {
	db *gorm.DB
}

func NewMerchantDocumentQueryRepository(db *gorm.DB) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{db: db}
}

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error) {
	offset := (req.Page - 1) * req.PageSize
	var docs []*models.MerchantDocument
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("document_type ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("document_id ASC").Find(&docs).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all merchant documents").WithInternal(err)
	}
	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindByIdDocument(ctx context.Context, id int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	err := r.db.WithContext(ctx).Where("document_id = ?", id).First(&doc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentQueryRepository) FindByActiveDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error) {
	offset := (req.Page - 1) * req.PageSize
	var docs []*models.MerchantDocument
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("document_type ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("document_id ASC").Find(&docs).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active merchant documents").WithInternal(err)
	}
	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindByTrashedDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*models.MerchantDocument, error) {
	offset := (req.Page - 1) * req.PageSize
	var docs []*models.MerchantDocument
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("document_type ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("document_id ASC").Find(&docs).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed merchant documents").WithInternal(err)
	}
	return docs, nil
}

func (r *merchantDocumentQueryRepository) CountAllDocuments(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("document_type ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *merchantDocumentQueryRepository) CountActiveDocuments(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("document_type ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *merchantDocumentQueryRepository) CountTrashedDocuments(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("document_type ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
