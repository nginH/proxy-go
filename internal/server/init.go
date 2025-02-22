package server

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/nginH/internal/repository/database"
)

type ProxyServer struct {
	*database.RedisClient
	proxy *httputil.ReverseProxy
}

func New() *ProxyServer {
	originURL := os.Getenv("ORIGIN_URL")
	if originURL == "" {
		panic("ORIGIN_URL is not set in the environment\n Exiting...")
	}

	backURL, err := url.Parse(originURL)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backURL)
	return &ProxyServer{
		RedisClient: database.NewRedisClient(),
		proxy:       proxy,
	}
}
