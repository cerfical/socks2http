package serv

import (
	"fmt"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"socks2http/internal/prox"
	"socks2http/internal/serv/http"
	"time"
)

func NewServer() *Server {
	return &Server{
		proxyAddr: addr.Addr{
			Scheme:   addr.Direct,
			Hostname: "localhost",
		},
		logger: log.NilLogger(),
	}
}

type Server struct {
	proxyAddr addr.Addr
	logger    log.Logger
	timeout   time.Duration
}

func (s *Server) SetLogger(logger log.Logger) {
	s.logger = logger
}

func (s *Server) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

func (s *Server) SetProxy(proxyAddr addr.Addr) {
	s.proxyAddr = proxyAddr
}

func (s *Server) Run(servAddr addr.Addr) error {
	s.logger.Info("starting server on %v", servAddr)
	if s.proxyAddr.Scheme != addr.Direct {
		s.logger.Info("using proxy %v", s.proxyAddr)
	}

	proxy, err := prox.NewProxy(s.proxyAddr, s.timeout)
	if err != nil {
		s.logger.Error("proxy disabled: %v", err)
		proxy = prox.Direct(s.timeout)
	}

	switch servAddr.Scheme {
	case addr.HTTP:
		return http.Run(servAddr.Host(), proxy, s.logger)
	default:
		return fmt.Errorf("unsupported server protocol scheme %q", servAddr.Scheme)
	}
}
