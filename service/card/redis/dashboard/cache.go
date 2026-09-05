package carddashboardmencache

import sharedcachehelpers "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"

type CardDashboardCache interface {
	CardDashboardTotalCache
	CardDashboardByCardNumberCache
}

type cardDashboardCaches struct {
	CardDashboardTotalCache
	CardDashboardByCardNumberCache
}

func NewMencacheDashboard(store *sharedcachehelpers.CacheStore) CardDashboardCache {
	return &cardDashboardCaches{
		CardDashboardTotalCache:        NewCardDashboardCache(store),
		CardDashboardByCardNumberCache: NewCardDashboardByCardNumberCache(store),
	}
}
