package serv

import (
	"fmt"
	"net"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox/cli"
	"github.com/cerfical/socks2http/internal/prox/serv/http"
)

func New(servAddr *addr.Addr, proxAddr *addr.Addr, timeout time.Duration, log *log.Logger) (*ProxyServer, error) {
	prox, err := cli.New(proxAddr, timeout)
	if err != nil {
		return nil, err
	}

	server := &ProxyServer{
		addr: servAddr,
		prox: prox,
		log:  log,
	}

	switch server.addr.Scheme {
	case addr.HTTP:
		server.handler = http.NewHandler(prox, log)
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %q", server.addr.Scheme)
	}

	return server, nil
}

type ProxyServer struct {
	addr    *addr.Addr
	prox    *cli.ProxyClient
	log     *log.Logger
	handler RequestHandler
}

type RequestHandler interface {
	Handle(cliConn net.Conn)
}

func (s *ProxyServer) Run() error {
	s.log.Infof("starting a server on %v", s.addr)
	if proxyAddr := s.prox.Addr(); proxyAddr.Scheme != addr.Direct {
		s.log.Infof("using a proxy %v", proxyAddr)
	} else {
		s.log.Infof("not using a proxy")
	}

	listener, err := net.Listen("tcp", s.addr.Host())
	if err != nil {
		return err
	}

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			s.log.Errorf("opening a client connection: %v", err)
			continue
		}

		go func() {
			defer func() {
				if err := cliConn.Close(); err != nil {
					s.log.Errorf("closing a client connection: %v", err)
				}
			}()
			s.handler.Handle(cliConn)
		}()
	}
}

func (s *ProxyServer) Addr() *addr.Addr {
	return s.addr
}
