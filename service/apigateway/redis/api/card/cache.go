package card_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type CardMencache interface {
	CardQueryCache
	CardCommandCache
}

type cardmencache struct {
	CardQueryCache
	CardCommandCache
}

func NewCardMencache(cacheStore *cache.CacheStore) CardMencache {

	return &cardmencache{
		CardCommandCache: NewCardCommandCache(cacheStore),
		CardQueryCache:   NewCardQueryCache(cacheStore),
	}
}
