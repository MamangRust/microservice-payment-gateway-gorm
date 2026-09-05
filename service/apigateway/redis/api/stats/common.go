package stats_cache

import "time"

const (
	StatsCachePrefix = "stats:"
	ttlDefault       = 5 * time.Minute
)
