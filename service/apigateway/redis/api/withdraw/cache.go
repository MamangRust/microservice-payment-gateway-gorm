package withdraw_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type WithdrawMencache interface {
	WithdrawQueryCache
	WithdrawCommandCache
}

type withdrawmencache struct {
	WithdrawQueryCache
	WithdrawCommandCache
}

func NewWithdrawMencache(cacheStore *cache.CacheStore) WithdrawMencache {
	return &withdrawmencache{
		WithdrawQueryCache:   NewWithdrawQueryCache(cacheStore),
		WithdrawCommandCache: NewWithdrawCommandCache(cacheStore),
	}
}
