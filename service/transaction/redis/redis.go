package mencache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type Mencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

// Mencache represents a cache store for transaction queries and commands.
type mencache struct {
	TransactionQueryCache
	TransactionCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		TransactionQueryCache:   NewTransactionQueryCache(cacheStore),
		TransactionCommandCache: NewTransactionCommandCache(cacheStore),
	}
}
