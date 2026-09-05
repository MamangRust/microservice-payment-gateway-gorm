package transfer_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type TransferMencache interface {
	TransferQueryCache
	TransferCommandCache
}

type transfermencache struct {
	TransferQueryCache
	TransferCommandCache
}

func NewTransferMencache(cacheStore *cache.CacheStore) TransferMencache {
	return &transfermencache{
		TransferQueryCache:   NewTransferQueryCache(cacheStore),
		TransferCommandCache: NewTransferCommandCache(cacheStore),
	}
}
