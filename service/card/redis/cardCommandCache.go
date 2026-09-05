package mencache

import (
	"context"
	"fmt"

	sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type cardCommandCache struct {
	store *sharedcachehelpers.CacheStore
}

func NewCardCommandCache(store *sharedcachehelpers.CacheStore) CardCommandCache {
	return &cardCommandCache{store: store}
}

func (c *cardCommandCache) DeleteCardCommandCache(ctx context.Context, id int) {
	key := fmt.Sprintf(cardByIdCacheKey, id)

	sharedcachehelpers.DeleteFromCache(ctx, c.store, key)
}
