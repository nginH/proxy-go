package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/standard" // Add this import for standard handlers
	"github.com/nginH/config"
	"github.com/nginH/internal/repository/database"
	"github.com/nginH/internal/service"
	logs "github.com/nginH/pkg/log"
	"gopkg.in/yaml.v3"
)

type ProxyServer struct {
	*database.RedisClient
	cacheService *service.CacheService
	cfg          *config.Config
	app          *caddy.Config
}

func New() *ProxyServer {
	cfg := &config.Config{}
	if err := loadConfig(cfg); err != nil {
		logs.Fatal("Failed to load config:", err)
	}
	redisClient := database.NewRedisClient()
	cacheService := service.NewCacheService(redisClient, cfg)

	app := &caddy.Config{
		Admin: &caddy.AdminConfig{
			Disabled: false,
		},
		Logging: &caddy.Logging{
			Sink: &caddy.SinkLog{},
		},
	}

	return &ProxyServer{
		RedisClient:  redisClient,
		cacheService: cacheService,
		cfg:          cfg,
		app:          app,
	}
}

func (s *ProxyServer) Start() error {
	if os.Getenv("ORIGIN_URL") == "" {
		return fmt.Errorf("ORIGIN_URL is not set")
	}

	// Create cache middleware configuration with proper initialization
	cacheHandlerRaw := json.RawMessage(`{
		"handler": "cache",
		"redis": {
			"Client": null,
			"Ctx": null
		},
		"cfg": {}
	}`)

	// Create reverse proxy handler
	reverseProxyRaw := json.RawMessage(`{
		"handler": "reverse_proxy",
		"upstreams": [{
			"dial": "` + os.Getenv("ORIGIN_URL") + `"
		}],
		"headers": {
			"request": {
				"set": {
					"Host": ["{http.reverse_proxy.upstream.hostport}"]
				}
			}
		}
	}`)
	routeConfig := caddyhttp.Route{
		HandlersRaw: []json.RawMessage{
			cacheHandlerRaw,
			reverseProxyRaw,
		},
	}

	httpApp := &caddyhttp.App{
		Servers: map[string]*caddyhttp.Server{
			"srv0": {
				Listen: []string{":" + os.Getenv("PORT")},
				Routes: caddyhttp.RouteList{routeConfig},
			},
		},
	}

	httpAppRaw, err := json.Marshal(httpApp)
	if err != nil {
		return fmt.Errorf("failed to marshal http app: %v", err)
	}

	s.app.AppsRaw = caddy.ModuleMap{
		"http": httpAppRaw,
	}

	if err := caddy.Run(s.app); err != nil {
		return fmt.Errorf("failed to start Caddy: %v", err)
	}

	return nil
}

func (s *ProxyServer) Stop() error {
	return caddy.Stop()
}

func loadConfig(cfg *config.Config) error {
	file, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return err
	}

	return nil
}
