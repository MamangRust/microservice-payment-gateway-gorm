package mencache

import (
	"context"
	"fmt"

	sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

// merchantCommandCache is a struct that represents the cache store
type merchantCommandCache struct {
	store *sharedcachehelpers.CacheStore
}

// NewMerchantCommandCache returns a new instance of merchantCommandCache
func NewMerchantCommandCache(store *sharedcachehelpers.CacheStore) MerchantCommandCache {
	return &merchantCommandCache{store: store}
}

// DeleteCachedMerchant removes the cache entry associated with the specified merchant ID.
// It formats the cache key using the merchant ID and deletes the entry from the cache store.
func (s *merchantCommandCache) DeleteCachedMerchant(ctx context.Context, id int) {
	key := fmt.Sprintf(merchantByIdCacheKey, id)

	sharedcachehelpers.DeleteFromCache(ctx, s.store, key)
}
