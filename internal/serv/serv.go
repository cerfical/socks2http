package serv

import (
	"cmp"
	"fmt"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"socks2http/internal/prox"
	"socks2http/internal/serv/http"
	"time"
)

type Server struct {
	ProxyAddr addr.Addr
	Timeout   time.Duration
	Logger    log.Logger
}

func (s *Server) Run(servAddr *addr.Addr) error {
	s.Logger = cmp.Or(s.Logger, log.NilLogger)

	s.Logger.Info("starting server on %v", servAddr)
	if s.ProxyAddr.Scheme != addr.Direct {
		s.Logger.Info("using proxy %v", &s.ProxyAddr)
	}

	proxy := prox.Proxy{
		ProxyAddr: s.ProxyAddr,
		Timeout:   s.Timeout,
	}

	switch servAddr.Scheme {
	case addr.HTTP:
		return http.Run(servAddr, &proxy, s.Logger)
	default:
		return fmt.Errorf("unsupported server protocol scheme %q", servAddr.Scheme)
	}
}
