package mencache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

// Mencache is a struct that holds various cache interfaces
type Mencache struct {
	RoleCommandCache RoleCommandCache
	RoleQueryCache   RoleQueryCache
}

func NewMencache(cacheStore *cache.CacheStore) *Mencache {
	return &Mencache{
		RoleCommandCache: NewRoleCommandCache(cacheStore),
		RoleQueryCache:   NewRoleQueryCache(cacheStore),
	}
}
