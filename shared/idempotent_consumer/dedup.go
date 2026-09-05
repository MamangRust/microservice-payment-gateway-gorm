// Package idempotent_consumer provides a lightweight in-memory dedup store
// for Kafka consumers. It tracks processed event_ids with a configurable
// TTL so that replays (at-least-once delivery) are safe without requiring
// an external database.
package idempotent_consumer

import (
	"sync"
	"time"
)

// Dedup tracks which event IDs have already been processed.
type Dedup struct {
	mu   sync.RWMutex
	seen map[string]time.Time
	ttl  time.Duration
}

// New creates a Dedup with the given TTL. Entries older than TTL are
// evicted during Mark or on the periodic cleanup tick.
func New(ttl time.Duration) *Dedup {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	d := &Dedup{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
	go d.cleanupLoop()
	return d
}

// IsDuplicate returns true if the event has already been processed AND
// is still within the TTL window. It marks the event as processed
// atomically so concurrent consumers on the same instance are safe.
func (d *Dedup) IsDuplicate(eventID string) bool {
	if eventID == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	ts, ok := d.seen[eventID]
	if ok && time.Since(ts) < d.ttl {
		return true
	}
	d.seen[eventID] = time.Now()
	return false
}

// Mark explicitly stores an event ID without checking. Useful for
// pre-registering IDs from batch processing.
func (d *Dedup) Mark(eventID string) {
	if eventID == "" {
		return
	}
	d.mu.Lock()
	d.seen[eventID] = time.Now()
	d.mu.Unlock()
}

// Size returns the current number of tracked entries.
func (d *Dedup) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.seen)
}

func (d *Dedup) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		d.evictExpired()
	}
}

func (d *Dedup) evictExpired() {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-d.ttl)
	for id, ts := range d.seen {
		if ts.Before(cutoff) {
			delete(d.seen, id)
		}
	}
}
