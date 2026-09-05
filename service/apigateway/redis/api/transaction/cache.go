package transaction_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type TransactionMencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

type transactionmencache struct {
	TransactionQueryCache
	TransactionCommandCache
}

func NewTransactionMencache(cacheStore *cache.CacheStore) TransactionMencache {
	return &transactionmencache{
		TransactionQueryCache:   NewTransactionQueryCache(cacheStore),
		TransactionCommandCache: NewTransactionCommandCache(cacheStore),
	}
}
