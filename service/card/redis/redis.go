package mencache

import (
	carddashboardmencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/redis/dashboard"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/redis/go-redis/v9"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

type Mencache interface {
	CardQueryCache
	CardCommandCache
	carddashboardmencache.CardDashboardCache
}

// Mencache is a struct that represents the cache store
type mencache struct {
	CardQueryCache
	CardCommandCache
	carddashboardmencache.CardDashboardCache
}

// Deps is a struct that represents the dependencies needed to create a Mencache
type Deps struct {
	Redis   redis.UniversalClient
	Logger  logger.LoggerInterface
	Metrics observability.CacheMetricsInterface
}

func NewMencache(cacheStore *cache.CacheStore) Mencache {
	return &mencache{
		CardCommandCache:   NewCardCommandCache(cacheStore),
		CardQueryCache:     NewCardQueryCache(cacheStore),
		CardDashboardCache: carddashboardmencache.NewMencacheDashboard(cacheStore),
	}
}
