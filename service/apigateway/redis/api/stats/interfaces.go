package stats_cache

import (
	"context"
)

type StatsCache interface {
	GetCache(ctx context.Context, key string) (interface{}, bool)
	SetCache(ctx context.Context, key string, data interface{})
	DeleteCache(ctx context.Context, key string)
}
