package stats_cache

import (
	"context"
	"sync"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

// flightWaitTimeout bounds how long a concurrent request waits for the
// in-flight "leader" that is computing a cache miss before it takes over and
// computes the value itself. Normal backends respond in well under a second,
// so this only fires when the leader errors out or hangs without populating
// the cache.
const flightWaitTimeout = 5 * time.Second

// flightCall tracks an in-flight computation for a single cache key.
// done is closed when the leader finishes (SetCache/DeleteCache), which
// unblocks every waiter for the same key.
type flightCall struct {
	done chan struct{}
}

type statsCache struct {
	store *cache.CacheStore

	mu       sync.Mutex
	inflight map[string]*flightCall
}

func NewStatsCache(store *cache.CacheStore) StatsCache {
	return &statsCache{
		store:    store,
		inflight: make(map[string]*flightCall),
	}
}

func (c *statsCache) GetCache(ctx context.Context, key string) (interface{}, bool) {
	for {
		if result, found := c.read(ctx, key); found {
			return result, true
		}

		// Cache miss: coalesce concurrent misses for the same key so only one
		// request (the "leader") computes and populates the cache; the rest
		// wait for it instead of stampeding the backend. This is an in-process
		// singleflight scoped to the shared cache-miss path.
		c.mu.Lock()
		fc, exists := c.inflight[key]
		if !exists {
			fc = &flightCall{done: make(chan struct{})}
			c.inflight[key] = fc
		}
		c.mu.Unlock()

		if !exists {
			// I am the leader: return a miss so the caller computes the value.
			return nil, false
		}

		// Waiter: block until the leader populates the cache (or times out).
		select {
		case <-fc.done:
			// Leader finished (SetCache/DeleteCache): loop and re-check the
			// cache; if it is still missing the leader failed to populate it
			// and we fall through to becoming the leader ourselves.
		case <-time.After(flightWaitTimeout):
			// The leader never populated the cache (backend error/hang). Take
			// over, but serialize the takeover so only one waiter becomes the
			// new leader instead of every waiter stampeding the backend.
			// fc is the entry fetched at the top of this iteration. Take over only
			// if we still own it; if another waiter already replaced it, loop and
			// wait on the fresh entry instead.
			c.mu.Lock()
			if current, ok := c.inflight[key]; ok && current == fc {
				c.inflight[key] = &flightCall{done: make(chan struct{})}
				c.mu.Unlock()
				return nil, false
			}
			c.mu.Unlock()
		case <-ctx.Done():
			return nil, false
		}
	}
}

func (c *statsCache) SetCache(ctx context.Context, key string, data interface{}) {
	if data == nil {
		// Intentionally do NOT finish(key) here: leaving the entry in flight
		// makes concurrent waiters take over serially instead of all becoming
		// leaders at once (which would re-create a stampede). The leader (this
		// caller) handles the value itself.
		return
	}
	cache.SetToCache(ctx, c.store, key, &data, ttlDefault)
	c.finish(key)
}

func (c *statsCache) DeleteCache(ctx context.Context, key string) {
	// Note: finish wakes every waiter, who then each re-register as leaders.
	// That is intentional for the invalidation path — coalescing only applies
	// to read-miss stampedes, and correctness wins here (data was just deleted).
	cache.DeleteFromCache(ctx, c.store, key)
	c.finish(key)
}

// finish closes the in-flight channel for key (if any) and removes the entry,
// unblocking every waiter. It is safe to call even when no computation is in
// flight for the key.
func (c *statsCache) finish(key string) {
	c.mu.Lock()
	if fc, ok := c.inflight[key]; ok {
		close(fc.done)
		delete(c.inflight, key)
	}
	c.mu.Unlock()
}

func (c *statsCache) read(ctx context.Context, key string) (interface{}, bool) {
	result, found := cache.GetFromCache[interface{}](ctx, c.store, key)
	if !found || result == nil {
		return nil, false
	}
	return *result, true
}
