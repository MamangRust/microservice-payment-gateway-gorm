package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type merchantCachedResponseAll struct {
	Data         []*models.Merchant `json:"data"`
	TotalRecords *int               `json:"total_records"`
}

type merchantQueryCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewMerchantQueryCache(store *sharedcachehelpers.CacheStore) MerchantQueryCache {
	return &merchantQueryCache{store: store}
}

func (m *merchantQueryCache) GetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool) {
	key := fmt.Sprintf(merchantAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[merchantCachedResponseAll](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *merchantQueryCache) SetCachedMerchants(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Merchant{}
	}
	key := fmt.Sprintf(merchantAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &merchantCachedResponseAll{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantQueryCache) GetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool) {
	key := fmt.Sprintf(merchantActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[merchantCachedResponseAll](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *merchantQueryCache) SetCachedMerchantActive(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Merchant{}
	}
	key := fmt.Sprintf(merchantActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &merchantCachedResponseAll{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantQueryCache) GetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, *int, bool) {
	key := fmt.Sprintf(merchantTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[merchantCachedResponseAll](ctx, m.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (m *merchantQueryCache) SetCachedMerchantTrashed(ctx context.Context, req *requests.FindAllMerchants, data []*models.Merchant, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Merchant{}
	}
	key := fmt.Sprintf(merchantTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &merchantCachedResponseAll{Data: data, TotalRecords: total}
	sharedcachehelpers.SetToCache(ctx, m.store, key, payload, ttlDefault)
}

func (m *merchantQueryCache) GetCachedMerchant(ctx context.Context, id int) (*models.Merchant, bool) {
	key := fmt.Sprintf(merchantByIdCacheKey, id)
	result, found := sharedcachehelpers.GetFromCache[*models.Merchant](ctx, m.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (m *merchantQueryCache) SetCachedMerchant(ctx context.Context, data *models.Merchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByIdCacheKey, data.MerchantID)
	sharedcachehelpers.SetToCache(ctx, m.store, key, data, ttlDefault)
}

func (m *merchantQueryCache) GetCachedMerchantsByUserId(ctx context.Context, userId int) ([]*models.Merchant, bool) {
	key := fmt.Sprintf(merchantByUserIdCacheKey, userId)
	result, found := sharedcachehelpers.GetFromCache[[]*models.Merchant](ctx, m.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (m *merchantQueryCache) SetCachedMerchantsByUserId(ctx context.Context, userId int, data []*models.Merchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByUserIdCacheKey, userId)
	sharedcachehelpers.SetToCache(ctx, m.store, key, &data, ttlDefault)
}

func (m *merchantQueryCache) GetCachedMerchantByApiKey(ctx context.Context, apiKey string) (*models.Merchant, bool) {
	key := fmt.Sprintf(merchantByApiKeyCacheKey, apiKey)
	result, found := sharedcachehelpers.GetFromCache[*models.Merchant](ctx, m.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (m *merchantQueryCache) SetCachedMerchantByApiKey(ctx context.Context, apiKey string, data *models.Merchant) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(merchantByApiKeyCacheKey, apiKey)
	sharedcachehelpers.SetToCache(ctx, m.store, key, data, ttlDefault)
}
