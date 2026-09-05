package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type withdrawCachedResponseAll struct {
	Data         []*repository.WithdrawQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type withdrawCachedResponseByCard struct {
	Data         []*repository.WithdrawQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type withdrawCachedResponseActive struct {
	Data         []*repository.WithdrawQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type withdrawCachedResponseTrashed struct {
	Data         []*repository.WithdrawQueryResult `json:"data"`
	TotalRecords *int                              `json:"total_records"`
}

type withdrawQueryCache struct {
	store *cache.CacheStore
}

func NewWithdrawQueryCache(store *cache.CacheStore) WithdrawQueryCache {
	return &withdrawQueryCache{store: store}
}

func (w *withdrawQueryCache) GetCachedWithdrawsCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool) {
	key := fmt.Sprintf(withdrawAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[withdrawCachedResponseAll](ctx, w.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (w *withdrawQueryCache) SetCachedWithdrawsCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.WithdrawQueryResult{}
	}

	key := fmt.Sprintf(withdrawAllCacheKey, req.Page, req.PageSize, req.Search)
	payload := &withdrawCachedResponseAll{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, w.store, key, payload, ttlDefault)
}

func (w *withdrawQueryCache) GetCachedWithdrawByCardCache(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, *int, bool) {
	key := fmt.Sprintf(withdrawByCardCacheKey, req.CardNumber, req.Page, req.PageSize, req.Search)

	result, found := cache.GetFromCache[withdrawCachedResponseByCard](ctx, w.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (w *withdrawQueryCache) SetCachedWithdrawByCardCache(ctx context.Context, req *requests.FindAllWithdrawCardNumber, data []*repository.WithdrawQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.WithdrawQueryResult{}
	}

	key := fmt.Sprintf(withdrawByCardCacheKey, req.CardNumber, req.Page, req.PageSize, req.Search)
	payload := &withdrawCachedResponseByCard{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, w.store, key, payload, ttlDefault)
}

func (w *withdrawQueryCache) GetCachedWithdrawActiveCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool) {
	key := fmt.Sprintf(withdrawActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[withdrawCachedResponseActive](ctx, w.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (w *withdrawQueryCache) SetCachedWithdrawActiveCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.WithdrawQueryResult{}
	}

	key := fmt.Sprintf(withdrawActiveCacheKey, req.Page, req.PageSize, req.Search)
	payload := &withdrawCachedResponseActive{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, w.store, key, payload, ttlDefault)
}

func (w *withdrawQueryCache) GetCachedWithdrawTrashedCache(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, bool) {
	key := fmt.Sprintf(withdrawTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := cache.GetFromCache[withdrawCachedResponseTrashed](ctx, w.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (w *withdrawQueryCache) SetCachedWithdrawTrashedCache(ctx context.Context, req *requests.FindAllWithdraws, data []*repository.WithdrawQueryResult, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*repository.WithdrawQueryResult{}
	}

	key := fmt.Sprintf(withdrawTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &withdrawCachedResponseTrashed{Data: data, TotalRecords: total}
	cache.SetToCache(ctx, w.store, key, payload, ttlDefault)
}

func (w *withdrawQueryCache) GetCachedWithdrawCache(ctx context.Context, id int) (*models.Withdraw, bool) {
	key := fmt.Sprintf(withdrawByIdCacheKey, id)
	result, found := cache.GetFromCache[*models.Withdraw](ctx, w.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (w *withdrawQueryCache) SetCachedWithdrawCache(ctx context.Context, data *models.Withdraw) {
	if data == nil {
		return
	}

	key := fmt.Sprintf(withdrawByIdCacheKey, data.WithdrawID)
	cache.SetToCache(ctx, w.store, key, data, ttlDefault)
}
