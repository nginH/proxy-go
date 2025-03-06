package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nginH/config"
	"github.com/nginH/internal/repository/database"
)

type CacheService struct {
	redis *database.RedisClient
	cfg   *config.Config
}

func NewCacheService(redis *database.RedisClient, cfg *config.Config) *CacheService {
	return &CacheService{
		redis: redis,
		cfg:   cfg,
	}
}

func (s *CacheService) GetCachedResponse(ctx context.Context, method, path string) (interface{}, error) {
	var route *config.RouteConfig
	for _, r := range s.cfg.Routes {
		if r.Query.Method == method && r.Query.Path == path {
			route = &r
			break
		}
	}

	if route == nil {
		return nil, fmt.Errorf("no matching route found")
	}

	cacheKey := fmt.Sprintf("%s:%s:%s", route.Name, method, path)

	// Try to get from cache
	data, err := s.redis.Get(cacheKey)
	if err != nil {
		return nil, err
	}

	var entry config.CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, err
	}

	return entry.Data, nil
}

func (s *CacheService) SetCachedResponse(ctx context.Context, method, path string, data interface{}, ttl time.Duration) error {
	// Find matching route
	var route *config.RouteConfig
	for _, r := range s.cfg.Routes {
		if r.Query.Method == method && r.Query.Path == path {
			route = &r
			break
		}
	}

	if route == nil {
		return fmt.Errorf("no matching route found")
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s:%s:%s", route.Name, method, path)

	// Create cache entry
	entry := config.CacheEntry{
		Data:      data,
		Timestamp: time.Now().Unix(),
		TTL:       int64(ttl.Seconds()),
	}

	// Marshal entry
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Set in Redis with TTL
	return s.redis.SetWithExpiration(cacheKey, string(jsonData), ttl)
}

func (s *CacheService) InvalidateCache(ctx context.Context, method, path string) error {
	// Find matching route and its invalidation rules
	var route *config.RouteConfig
	for _, r := range s.cfg.Routes {
		for _, inv := range r.Invalidate {
			if inv.Method == method && inv.Path == path {
				route = &r
				break
			}
		}
		if route != nil {
			break
		}
	}

	if route == nil {
		return fmt.Errorf("no matching invalidation rule found")
	}

	// Generate cache key for the query route
	cacheKey := fmt.Sprintf("%s:%s:%s", route.Name, route.Query.Method, route.Query.Path)

	// Delete from cache
	return s.redis.Del(cacheKey)
}
