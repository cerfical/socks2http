package serv

import (
	"fmt"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox/cli"
	"github.com/cerfical/socks2http/internal/prox/serv/http"
)

func New(servAddr *addr.Addr, proxyAddr *addr.Addr, timeout time.Duration, logger *log.Logger) (*Server, error) {
	proxy, err := cli.New(proxyAddr, timeout)
	if err != nil {
		return nil, err
	}

	server := &Server{
		addr:   servAddr,
		proxy:  proxy,
		logger: logger,
	}

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
	proxy  *cli.Proxy
	logger *log.Logger
	run    func() error
}

func (s *Server) Addr() *addr.Addr {
	return s.addr
}

func (s *Server) Run() error {
	s.logger.Infof("starting server on %v", s.Addr())
	if proxyAddr := s.proxy.Addr(); proxyAddr.Scheme != addr.Direct {
		s.logger.Infof("using proxy %v", proxyAddr)
	} else {
		s.logger.Infof("not using proxy")
	}
	return s.run()
}
