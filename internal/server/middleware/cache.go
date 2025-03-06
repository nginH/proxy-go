package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nginH/internal/service"
)

func init() {
	caddy.RegisterModule(CacheMiddleware{})
	httpcaddyfile.RegisterHandlerDirective("cache", parseCaddyfile)
}

type CacheMiddleware struct {
	CacheService *service.CacheService
}

func (CacheMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.cache",
		New: func() caddy.Module { return new(CacheMiddleware) },
	}
}

func (m *CacheMiddleware) Provision(ctx caddy.Context) error {
	// Get cache service from context
	app, err := ctx.App("cache")
	if err != nil {
		return err
	}
	m.CacheService = app.(*service.CacheService)
	return nil
}

func (m *CacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Check if request should be cached
	cachedData, err := m.CacheService.GetCachedResponse(r.Context(), r.Method, r.URL.Path)
	if err == nil {
		// Return cached data
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(cachedData)
	}

	// Create response recorder
	var buf bytes.Buffer
	shouldBuffer := func(status int, header http.Header) bool { return true }
	rec := caddyhttp.NewResponseRecorder(w, &buf, shouldBuffer)

	// Call next handler
	if err := next.ServeHTTP(rec, r); err != nil {
		return err
	}

	// Cache the response
	if rec.Status() == http.StatusOK {
		var data interface{}
		if err := json.NewDecoder(&buf).Decode(&data); err != nil {
			return err
		}

		// Set cache with 5 minutes TTL
		if err := m.CacheService.SetCachedResponse(r.Context(), r.Method, r.URL.Path, data, 5*time.Minute); err != nil {
			return err
		}
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	return &CacheMiddleware{}, nil
}
