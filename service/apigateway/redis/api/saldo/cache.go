package saldo_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type SaldoMencache interface {
	SaldoQueryCache
	SaldoCommandCache
}

type saldomencache struct {
	SaldoQueryCache
	SaldoCommandCache
}

func NewSaldoMencache(cacheStore *cache.CacheStore) SaldoMencache {
	return &saldomencache{
		SaldoQueryCache:   NewSaldoQueryCache(cacheStore),
		SaldoCommandCache: NewSaldoCommandCache(cacheStore),
	}
}
