package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nginH/config"
	"github.com/nginH/internal/repository/database"
	"github.com/nginH/internal/server/middleware"
	"github.com/nginH/internal/service"
	logs "github.com/nginH/pkg/log"
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
			Disabled: true,
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
	routeConfig := caddyhttp.Route{
		HandlersRaw: []json.RawMessage{
			json.RawMessage(`{
				"handler": "reverse_proxy",
				"upstreams": [{"dial": "` + os.Getenv("ORIGIN_URL") + `"}]
			}`),
		},
	}

	cacheHandler := &middleware.CacheMiddleware{
		CacheService: s.cacheService,
	}
	cacheHandlerRaw, err := json.Marshal(cacheHandler)
	if err != nil {
		return fmt.Errorf("failed to marshal cache handler: %v", err)
	}

	routeConfig.HandlersRaw = append([]json.RawMessage{
		cacheHandlerRaw,
	}, routeConfig.HandlersRaw...)

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

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return err
	}

	return nil
}
