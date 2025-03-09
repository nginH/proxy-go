package cache_middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nginH/config"
	"github.com/nginH/internal/repository/database"
	"github.com/nginH/internal/service"
	logs "github.com/nginH/pkg/log"
)

func init() {
	caddy.RegisterModule(CacheMiddleware{})
	httpcaddyfile.RegisterHandlerDirective("cache", parseCaddyfile)
}

type CacheMiddleware struct {
	CacheService *service.CacheService `json:"CacheService,omitempty"`
	Redis        *database.RedisClient `json:"redis,omitempty"`
	Config       *config.Config        `json:"cfg,omitempty"`
}

func (CacheMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.cache",
		New: func() caddy.Module { return new(CacheMiddleware) },
	}
}

func (m *CacheMiddleware) Provision(ctx caddy.Context) error {
	logs.Info("Provisioning cache middleware")

	// If Redis is not provided, create a new client
	if m.Redis == nil {
		m.Redis = database.NewRedisClient()
	}

	// If config is not provided, load it
	if m.Config == nil {
		m.Config = &config.Config{}
		// Load default config or minimal required settings
		// You might want to implement a proper way to load config here
	}

	// Initialize the cache service
	m.CacheService = service.NewCacheService(m.Redis, m.Config)
	logs.Info("Cache middleware provisioned successfully")

	return nil
}

func (m *CacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Skip non-GET requests for caching
	if r.Method != http.MethodGet {
		return next.ServeHTTP(w, r)
	}

	// Check if request should be cached
	cachedData, err := m.CacheService.GetCachedResponse(r.Context(), r.Method, r.URL.Path)
	if err == nil {
		// Return cached data
		logs.Info("Cache hit for", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		return json.NewEncoder(w).Encode(cachedData)
	}

	logs.Info("Cache miss for", r.URL.Path)

	// Create response recorder
	buf := &bytes.Buffer{}
	rec := caddyhttp.NewResponseRecorder(w, buf, func(status int, header http.Header) bool {
		return status == http.StatusOK && header.Get("Content-Type") == "application/json"
	})

	// Call next handler
	if err := next.ServeHTTP(rec, r); err != nil {
		return err
	}

	// Only cache 200 OK responses
	if rec.Status() == http.StatusOK {
		// Check if response is valid JSON
		bodyBytes := buf.Bytes()
		if json.Valid(bodyBytes) {
			var data interface{}
			if err := json.Unmarshal(bodyBytes, &data); err == nil {
				// Set cache with 5 minutes TTL
				if err := m.CacheService.SetCachedResponse(r.Context(), r.Method, r.URL.Path, data, 5*time.Minute); err != nil {
					logs.Error("Failed to cache response:", err)
				} else {
					logs.Info("Cached response for", r.URL.Path)
				}
			}
		}

		// Write the response body to the original writer
		w.Header().Set("X-Cache", "MISS")
		_, err = io.Copy(w, bytes.NewReader(bodyBytes))
		return err
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m CacheMiddleware

	// No configuration options for now, but you could add them here
	if err := h.Dispenser.NextArg(); err {
		return nil, h.Dispenser.ArgErr()
	}

	return &m, nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*CacheMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*CacheMiddleware)(nil)
)
