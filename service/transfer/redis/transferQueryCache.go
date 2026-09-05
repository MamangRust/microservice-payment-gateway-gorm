package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type transferCacheResponseAll struct {
	Data         []*repository.TransferQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type transferCacheResponseActive struct {
	Data         []*repository.TransferQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type transferCacheResponseTrashed struct {
	Data         []*repository.TransferQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type transferQueryCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewTransferQueryCache(store *sharedcachehelpers.CacheStore) TransferQueryCache {
	return &transferQueryCache{store: store}
}

func (c *transferQueryCache) GetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool) {
	key := fmt.Sprintf(transferAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[transferCacheResponseAll](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *transferQueryCache) SetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TransferQueryResult{}
	}

	key := fmt.Sprintf(transferAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transferCacheResponseAll{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *transferQueryCache) GetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool) {
	key := fmt.Sprintf(transferActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[transferCacheResponseActive](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (c *transferQueryCache) SetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TransferQueryResult{}
	}

	key := fmt.Sprintf(transferActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transferCacheResponseActive{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *transferQueryCache) GetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool) {
	key := fmt.Sprintf(transferTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := sharedcachehelpers.GetFromCache[transferCacheResponseTrashed](ctx, c.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (c *transferQueryCache) SetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.TransferQueryResult{}
	}

	key := fmt.Sprintf(transferTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &transferCacheResponseTrashed{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *transferQueryCache) GetCachedTransferCache(ctx context.Context, id int) (*models.Transfer, bool) {
	key := fmt.Sprintf(transferByIdCacheKey, id)
	result, found := sharedcachehelpers.GetFromCache[*models.Transfer](ctx, c.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (c *transferQueryCache) SetCachedTransferCache(ctx context.Context, data *models.Transfer) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferByIdCacheKey, data.TransferID)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *transferQueryCache) GetCachedTransferByFrom(ctx context.Context, from string) ([]*models.Transfer, bool) {
	key := fmt.Sprintf(transferByFromCacheKey, from)
	result, found := sharedcachehelpers.GetFromCache[[]*models.Transfer](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *transferQueryCache) SetCachedTransferByFrom(ctx context.Context, from string, data []*models.Transfer) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferByFromCacheKey, from)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}

func (c *transferQueryCache) GetCachedTransferByTo(ctx context.Context, to string) ([]*models.Transfer, bool) {
	key := fmt.Sprintf(transferByToCacheKey, to)

	result, found := sharedcachehelpers.GetFromCache[[]*models.Transfer](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *transferQueryCache) SetCachedTransferByTo(ctx context.Context, to string, data []*models.Transfer) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(transferByToCacheKey, to)
	sharedcachehelpers.SetToCache(ctx, c.store, key, &data, ttlDefault)
}
