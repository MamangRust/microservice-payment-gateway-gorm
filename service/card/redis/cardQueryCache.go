package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type cardCachedResponse struct {
	Data         []*models.Card `json:"data"`
	TotalRecords *int           `json:"total_records"`
}

type cardQueryCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardQueryCache(store *sharedcachehelpers.CacheStore) CardQueryCache {
	return &cardQueryCache{store: store}
}

func (c *cardQueryCache) GetByIdCache(ctx context.Context, cardID int) (*models.Card, bool) {
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	result, found := sharedcachehelpers.GetFromCache[*models.Card](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetByUserIDCache(ctx context.Context, userID int) (*models.Card, bool) {
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	result, found := sharedcachehelpers.GetFromCache[*models.Card](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetByCardNumberCache(ctx context.Context, cardNumber string) (*models.Card, bool) {
	key := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)
	result, found := sharedcachehelpers.GetFromCache[*models.Card](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetUserCardByCardNumberCache(ctx context.Context, cardNumber string) (*models.Card, bool) {
	key := fmt.Sprintf(cardUserByCardNumCacheKey, cardNumber)
	result, found := sharedcachehelpers.GetFromCache[*models.Card](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}

func (c *cardQueryCache) GetFindAllCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool) {
	key := fmt.Sprintf(cardAllCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[cardCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *cardQueryCache) GetByActiveCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool) {
	key := fmt.Sprintf(cardActiveCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[cardCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *cardQueryCache) GetByTrashedCache(ctx context.Context, req *requests.FindAllCards) ([]*models.Card, *int, bool) {
	key := fmt.Sprintf(cardTrashedCacheKey, req.Page, req.PageSize, req.Search)
	result, found := sharedcachehelpers.GetFromCache[cardCachedResponse](ctx, c.store, key)
	if !found || result == nil {
		return nil, nil, false
	}
	return result.Data, result.TotalRecords, true
}

func (c *cardQueryCache) SetByIdCache(ctx context.Context, cardID int, data *models.Card) {
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByUserIDCache(ctx context.Context, userID int, data *models.Card) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetByCardNumberCache(ctx context.Context, cardNumber string, data *models.Card) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) SetFindAllCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Card{}
	}
	payload := &cardCachedResponse{Data: data, TotalRecords: total}
	key := fmt.Sprintf(cardAllCacheKey, req.Page, req.PageSize, req.Search)
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *cardQueryCache) SetByActiveCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Card{}
	}
	payload := &cardCachedResponse{Data: data, TotalRecords: total}
	key := fmt.Sprintf(cardActiveCacheKey, req.Page, req.PageSize, req.Search)
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *cardQueryCache) SetByTrashedCache(ctx context.Context, req *requests.FindAllCards, data []*models.Card, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}
	if data == nil {
		data = []*models.Card{}
	}
	payload := &cardCachedResponse{Data: data, TotalRecords: total}
	key := fmt.Sprintf(cardTrashedCacheKey, req.Page, req.PageSize, req.Search)
	sharedcachehelpers.SetToCache(ctx, c.store, key, payload, ttlDefault)
}

func (c *cardQueryCache) SetUserCardByCardNumberCache(ctx context.Context, cardNumber string, data *models.Card) {
	if data == nil {
		return
	}
	key := fmt.Sprintf(cardUserByCardNumCacheKey, cardNumber)
	sharedcachehelpers.SetToCache(ctx, c.store, key, data, ttlDefault)
}

func (c *cardQueryCache) DeleteByIdCache(ctx context.Context, cardID int) {
	key := fmt.Sprintf(cardByIdCacheKey, cardID)
	sharedcachehelpers.DeleteFromCache(ctx, c.store, key)
}

func (c *cardQueryCache) DeleteByUserIDCache(ctx context.Context, userID int) {
	key := fmt.Sprintf(cardByUserIdCacheKey, userID)
	sharedcachehelpers.DeleteFromCache(ctx, c.store, key)
}

func (c *cardQueryCache) DeleteByCardNumberCache(ctx context.Context, cardNumber string) {
	cardKey := fmt.Sprintf(cardByCardNumCacheKey, cardNumber)
	userCardKey := fmt.Sprintf(cardUserByCardNumCacheKey, cardNumber)
	sharedcachehelpers.DeleteFromCache(ctx, c.store, cardKey)
	sharedcachehelpers.DeleteFromCache(ctx, c.store, userCardKey)
}
