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

func NewServer(servAddr *addr.Addr, proxyAddr *addr.Addr, timeout time.Duration, logger log.Logger) (*Server, error) {
	proxy, err := prox.NewProxy(proxyAddr, timeout)
	if err != nil {
		return nil, err
	}

	server := &Server{
		addr:   servAddr,
		proxy:  proxy,
		logger: cmp.Or(logger, log.NilLogger),
	}

	// set defaults if no proper values are provided
	server.addr.Scheme = cmp.Or(server.addr.Scheme, addr.HTTP)
	server.addr.Port = cmp.Or(server.addr.Port, server.addr.Scheme.Port())

	switch server.addr.Scheme {
	case addr.HTTP:
		server.run = func() error {
			return http.Run(server.addr, proxy, logger)
		}
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %q", server.addr.Scheme)
	}

	return server, nil
}

type Server struct {
	addr   *addr.Addr
	proxy  *prox.Proxy
	logger log.Logger
	run    func() error
}

func (s *Server) Addr() *addr.Addr {
	return s.addr
}

func (s *Server) Run() error {
	s.logger.Info("starting server on %v", s.Addr())
	if proxyAddr := s.proxy.Addr(); proxyAddr.Scheme != addr.Direct {
		s.logger.Info("using proxy %v", proxyAddr)
	} else {
		s.logger.Info("not using proxy")
	}
	return s.run()
}
