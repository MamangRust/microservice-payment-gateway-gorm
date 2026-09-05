package role_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type RoleMencache interface {
	RoleCommandCache
	RoleQueryCache
}

type roleMencache struct {
	RoleCommandCache
	RoleQueryCache
}

func NewRoleMencache(cacheStore *cache.CacheStore) RoleMencache {
	return &roleMencache{
		RoleCommandCache: NewRoleCommandCache(cacheStore),
		RoleQueryCache:   NewRoleQueryCache(cacheStore),
	}
}
