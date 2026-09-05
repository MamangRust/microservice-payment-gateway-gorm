package mencache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type Mencache interface {
	TransferQueryCache
	TransferCommandCache
}

type mencache struct {
	TransferQueryCache
	TransferCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		TransferQueryCache:   NewTransferQueryCache(cacheStore),
		TransferCommandCache: NewTransferCommandCache(cacheStore),
	}
}
