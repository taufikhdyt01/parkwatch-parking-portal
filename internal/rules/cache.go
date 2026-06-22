package rules

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// activeCacheKey holds the JSON of the currently active rule version.
const activeCacheKey = "rules:active"

// Cache is a thin Redis cache for the active ruleset, which is read on every
// violation submission. A nil client disables caching (the service still works).
type Cache struct {
	rdb    *redis.Client
	ttl    time.Duration
	logger *slog.Logger
}

// NewCache builds a cache. rdb may be nil to disable caching.
func NewCache(rdb *redis.Client, ttl time.Duration, logger *slog.Logger) *Cache {
	return &Cache{rdb: rdb, ttl: ttl, logger: logger}
}

// GetActive returns the cached active version, or (nil, false) on miss/disabled.
func (c *Cache) GetActive(ctx context.Context) (*Version, bool) {
	if c.rdb == nil {
		return nil, false
	}
	data, err := c.rdb.Get(ctx, activeCacheKey).Bytes()
	if err != nil {
		return nil, false
	}
	var v Version
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, false
	}
	return &v, true
}

// SetActive caches the active version.
func (c *Cache) SetActive(ctx context.Context, v *Version) {
	if c.rdb == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := c.rdb.Set(ctx, activeCacheKey, data, c.ttl).Err(); err != nil {
		c.logger.Warn("cache set failed", "err", err)
	}
}

// Invalidate drops the cached active version (called after publishing).
func (c *Cache) Invalidate(ctx context.Context) {
	if c.rdb == nil {
		return
	}
	if err := c.rdb.Del(ctx, activeCacheKey).Err(); err != nil {
		c.logger.Warn("cache invalidate failed", "err", err)
	}
}
