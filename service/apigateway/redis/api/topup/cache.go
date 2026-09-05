package topup_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type TopupMencache interface {
	TopupQueryCache
	TopupCommandCache
}

type mencache struct {
	TopupQueryCache
	TopupCommandCache
}

func NewTopupMencache(cacheStore *cache.CacheStore) TopupMencache {
	return &mencache{
		TopupQueryCache:   NewTopupQueryCache(cacheStore),
		TopupCommandCache: NewTopupCommandCache(cacheStore),
	}
}
