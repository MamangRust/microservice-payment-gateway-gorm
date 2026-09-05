package merchant_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type mencache struct {
	MerchantQueryCache
	MerchantCommandCache
	MerchantTransactionCache
}

type MerchantMencache interface {
	MerchantQueryCache
	MerchantCommandCache
	MerchantTransactionCache
}

func NewMerchantMencache(cacheStore *cache.CacheStore) MerchantMencache {

	return &mencache{
		MerchantQueryCache:   NewMerchantQueryCache(cacheStore),
		MerchantCommandCache: NewMerchantCommandCache(cacheStore),

		MerchantTransactionCache: NewMerchantTransactionCache(cacheStore),
	}
}
