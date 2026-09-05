package transfer_cache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type transferCommandCache struct {
	store *cache.CacheStore
}

func NewTransferCommandCache(store *cache.CacheStore) TransferCommandCache {
	return &transferCommandCache{store: store}
}

func (t *transferCommandCache) DeleteTransferCache(ctx context.Context, id int) {
	cache.DeleteFromCache(ctx, t.store, fmt.Sprintf(transferByIdCacheKey, id))
}
