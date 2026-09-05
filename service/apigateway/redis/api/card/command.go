package card_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type cardCommandCache struct {
	store *cache.CacheStore
}

func NewCardCommandCache(store *cache.CacheStore) CardCommandCache {
	return &cardCommandCache{store: store}
}

func (c *cardCommandCache) DeleteCardCommandCache(ctx context.Context, id int) {
	key := fmt.Sprintf(cardByIdCacheKey, id)

	cache.DeleteFromCache(ctx, c.store, key)
}
