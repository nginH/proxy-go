package config

type Config struct {
	Routes []RouteConfig `yaml:"routes"`
}

type RouteConfig struct {
	Name       string           `yaml:"name"`
	Query      QueryRule        `yaml:"query"`
	Invalidate []InvalidateRule `yaml:"invalidate"`
}

type QueryRule struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type InvalidateRule struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	TTL       int64       `json:"ttl"`
}

type Service struct {
	Fetch string `json:"fetch"` // user to fetch request
}

// request
type Request struct {
	Path   string `json:"path"`   // path to fetch
	Body   string `json:"body"`   // body to fetch
	Params string `json:"params"` // params to fetch
	Query  string `json:"query"`  // query to fetch
}

// response
type Response struct {
	Body *string `json:"body"` // body to fetch
}
