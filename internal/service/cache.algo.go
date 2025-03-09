package service

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/nginH/config"
// 	"github.com/nginH/internal/repository/database"
// 	"github.com/nginH/pkg/cache"
// 	"google.golang.org/grpc"
// )

// // CacheService manages Redis caching with RL optimization
// type CacheService struct {
// 	redis      *database.RedisClient
// 	config     *config.Config
// 	routeRules map[string]RouteRule
// 	rlClient   cache.CacheServiceClient
// 	rlConn     *grpc.ClientConn
// 	mutex      sync.RWMutex
// }

// // RouteRule defines caching rules for a specific route
// type RouteRule struct {
// 	Query      RoutePattern   `json:"query"`
// 	Invalidate []RoutePattern `json:"invalidate"`
// 	TTL        int            `json:"ttl"`
// }

// // RoutePattern represents a route pattern
// type RoutePattern struct {
// 	Method string `json:"method"`
// 	Path   string `json:"path"`
// }

// // NewCacheService creates a new cache service
// func NewCacheService(redis *database.RedisClient, cfg *config.Config) *CacheService {
// 	service := &CacheService{
// 		redis:      redis,
// 		config:     cfg,
// 		routeRules: make(map[string]RouteRule),
// 	}

// 	// 	// Load initial route rules from config
// 	if cfg != nil && cfg.Routes != nil {
// 		for _, route := range cfg.Routes {
// 			key := fmt.Sprintf("%s:%s", route.Query.Method, route.Query.Path)
// 			service.routeRules[key] = route
// 		}
// 	}

// 	// 	// Initialize gRPC connection to RL service
// 	service.connectToRLService()

// 	return service
// }

// // connectToRLService establishes connection to the Python RL service
// func (s *CacheService) connectToRLService() {
// 	rlServiceAddr := s.config.RLServiceAddr
// 	if rlServiceAddr == "" {
// 		rlServiceAddr = "localhost:50051"
// 	}

// 	conn, err := grpc.Dial(rlServiceAddr, grpc.WithInsecure())
// 	if err != nil {
// 		logs.Error("Failed to connect to RL service:", err)
// 		return
// 	}

// 	s.rlConn = conn
// 	s.rlClient = cache.NewCacheServiceClient(conn)
// 	logs.Info("Connected to RL service at", rlServiceAddr)
// }

// // Close closes the gRPC connection
// func (s *CacheService) Close() {
// 	if s.rlConn != nil {
// 		s.rlConn.Close()
// 	}
// }

// // GetCachedResponse retrieves a cached response
// func (s *CacheService) GetCachedResponse(ctx context.Context, method, path string) (interface{}, error) {
// 	key := s.generateCacheKey(method, path)

// 	// Get from Redis
// 	data, err := s.redis.Client.Get(s.redis.Ctx, key).Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Unmarshal the cached data
// 	var result interface{}
// 	if err := json.Unmarshal([]byte(data), &result); err != nil {
// 		return nil, err
// 	}

// 	// Log cache hit to RL service
// 	go s.logRequest(method, path, true, 0) // TODO: Response time needed for LogRequest

// 	return result, nil
// }

// // // SetCachedResponse caches a response with dynamic TTL
// func (s *CacheService) SetCachedResponse(ctx context.Context, method, path string, data interface{}, defaultTTL time.Duration) error {
// 	// Get caching decision from RL model
// 	shouldCache, ttl := s.getCacheDecision(method, path, int(defaultTTL.Seconds()))

// 	if !shouldCache {
// 		return fmt.Errorf("RL model decided not to cache this route")
// 	}

// 	// Serialize data
// 	bytes, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}

// 	// Generate cache key
// 	key := s.generateCacheKey(method, path)

// 	// Store in Redis with TTL
// 	return s.redis.Client.Set(s.redis.Ctx, key, bytes, time.Duration(ttl)*time.Second).Err()
// }

// // InvalidateCache invalidates cache for a route and its related routes
// func (s *CacheService) InvalidateCache(ctx context.Context, method, path string) error {
// 	// Find all routes that should be invalidated
// 	routesToInvalidate := s.findRoutesToInvalidate(method, path)

// 	// Delete from Redis
// 	for _, route := range routesToInvalidate {
// 		key := s.generateCacheKey(route.Method, route.Path)
// 		if err := s.redis.Client.Del(s.redis.Ctx, key).Err(); err != nil {
// 			logs.Error("Failed to invalidate cache:", err)
// 		}
// 	}

// 	// 	// Inform RL service about invalidation
// 	if s.rlClient != nil {
// 		_, err := s.rlClient.InvalidateCache(ctx, &cache.RouteRequest{
// 			Method: method,
// 			Path:   path,
// 		})
// 		if err != nil {
// 			logs.Error("Failed to inform RL service about cache invalidation:", err)
// 		}
// 	}

// 	return nil
// }

// // // PreloadCache proactively caches popular routes
// func (s *CacheService) PreloadCache(ctx context.Context) error {
// 	if s.rlClient == nil {
// 		return fmt.Errorf("RL service client not initialized")
// 	}

// 	// Get popular routes from RL service
// 	resp, err := s.rlClient.GetPopularRoutes(ctx, &cache.PopularRoutesRequest{Threshold: 10}) // Example threshold
// 	if err != nil {
// 		return fmt.Errorf("failed to get popular routes from RL service: %w", err)
// 	}

// 	// 	// Iterate through popular routes and cache them (using default TTL or route-specific TTL)
// 	for _, route := range resp.GetRoutes() {
// 		// Fetch data for the route (you need to implement this based on your application logic)
// 		data, fetchErr := s.fetchDataForRoute(ctx, route.GetMethod(), route.GetPath()) // Implement fetchDataForRoute
// 		if fetchErr != nil {
// 			logs.Warnf("Failed to fetch data for popular route %s %s: %v", route.GetMethod(), route.GetPath(), fetchErr)
// 			continue
// 		}

// 		// Determine TTL (use default or route-specific rule if available)
// 		ttl := s.getDefaultTTLForRoute(route.GetMethod(), route.GetPath()) // Implement getDefaultTTLForRoute

// 		// Cache the data
// 		if err := s.SetCachedResponse(ctx, route.GetMethod(), route.GetPath(), data, ttl); err != nil {
// 			logs.Errorf("Failed to preload cache for route %s %s: %v", route.GetMethod(), route.GetPath(), err)
// 		} else {
// 			logs.Infof("Preloaded cache for popular route %s %s with TTL %s", route.GetMethod(), route.GetPath(), ttl)
// 		}
// 	}

// 	return nil
// }

// // // getCacheDecision queries the RL service for a caching decision
// func (s *CacheService) getCacheDecision(method, path string, defaultTTL int) (shouldCache bool, ttl int) {
// 	if s.rlClient == nil {
// 		logs.Warn("RL service client not initialized, using default caching behavior")
// 		return true, defaultTTL // Default to caching if RL service is not available
// 	}

// 	resp, err := s.rlClient.GetCacheDecision(s.redis.Ctx, &cache.RouteRequest{
// 		Method: method,
// 		Path:   path,
// 	})
// 	if err != nil {
// 		logs.Errorf("Failed to get cache decision from RL service: %v", err)
// 		return true, defaultTTL // Default to caching on error
// 	}

// 	return resp.GetShouldCache(), int(resp.GetTtl())
// }

// // // logRequest sends request log information to the RL service
// func (s *CacheService) logRequest(method, path string, cacheHit bool, responseTime float32) {
// 	if s.rlClient == nil {
// 		logs.Warn("RL service client not initialized, cannot log request")
// 		return
// 	}

// 	_, err := s.rlClient.LogRequest(s.redis.Ctx, &cache.RequestLog{
// 		Method:       method,
// 		Path:         path,
// 		CacheHit:     cacheHit,
// 		ResponseTime: responseTime, // Need to get actual response time here in proxy
// 		// ... other request details can be added if needed by RL model
// 	})
// 	if err != nil {
// 		logs.Errorf("Failed to log request to RL service: %v", err)
// 	}
// }

// // // generateCacheKey generates a cache key
// func (s *CacheService) generateCacheKey(method, path string) string {
// 	return fmt.Sprintf("cache:%s:%s", method, path)
// }

// // // findRoutesToInvalidate finds routes to invalidate based on route rules (example logic, needs to be refined)
// func (s *CacheService) findRoutesToInvalidate(method, path string) []RoutePattern {
// 	var routesToInvalidate []RoutePattern

// 	// 	// Example: Invalidate routes that have this path as part of their invalidation rule
// 	for _, rule := range s.routeRules {
// 		for _, invalidationRule := range rule.Invalidate {
// 			if invalidationRule.Path == path { // Simple path match, refine as needed
// 				routesToInvalidate = append(routesToInvalidate, rule.Query)
// 				break // Avoid adding the same route multiple times from different rules
// 			}
// 		}
// 	}
// 	return routesToInvalidate
// }

// // fetchDataForRoute -  **Needs Implementation**:  Implement logic to fetch actual data for a given route.
// // This is application-specific. It might involve calling backend services, databases, etc.
// func (s *CacheService) fetchDataForRoute(ctx context.Context, method string, path string) (interface{}, error) {
// 	// **PLACEHOLDER - Replace with actual data fetching logic for your application**
// 	// Example:  If you have a backend service to fetch data, call it here.
// 	// For now, returning a dummy string.
// 	dummyData := fmt.Sprintf("Data for route %s %s fetched proactively", method, path)
// 	return dummyData, nil
// }

// // getDefaultTTLForRoute - **Needs Implementation**: Implement logic to determine default TTL for a route.
// // This could be based on route rules, configuration, or a general default.
// func (s *CacheService) getDefaultTTLForRoute(method string, path string) time.Duration {
// 	// **PLACEHOLDER - Replace with actual TTL determination logic**
// 	// Example: Check route rules, or return a global default TTL.
// 	return 300 * time.Second // Default 5 minutes
// }
