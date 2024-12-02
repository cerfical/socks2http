package serv

import (
	"fmt"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"socks2http/internal/prox"
	"socks2http/internal/serv/http"
	"time"
)

type Server interface {
	Run() error
}

func NewServer(servAddr addr.Addr, proxyAddr addr.Addr, timeout time.Duration, logger log.Logger) (Server, error) {
	proxy, err := prox.NewProxy(proxyAddr, timeout)
	if err != nil {
		return nil, fmt.Errorf("proxy chaining: %v", err)
	}

	switch servAddr.Scheme {
	case addr.HTTP:
		return &http.HTTPServer{
			Host:   servAddr.Host(),
			Proxy:  proxy,
			Logger: logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %q", servAddr.Scheme)
	}
}
