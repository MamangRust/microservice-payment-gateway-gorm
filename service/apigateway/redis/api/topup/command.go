package topup_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type topupCommandCache struct {
	store *cache.CacheStore
}

func NewTopupCommandCache(store *cache.CacheStore) TopupCommandCache {
	return &topupCommandCache{store: store}
}

func (c *topupCommandCache) DeleteCachedTopupCache(ctx context.Context, id int) {
	key := fmt.Sprintf(topupByIdCacheKey, id)
	cache.DeleteFromCache(ctx, c.store, key)
}
