package serv

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/cli"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/serv/http"
)

func New(servAddr, proxAddr *addr.Addr, timeout time.Duration, log *log.Logger) (*ProxyServer, error) {
	prox, err := cli.New(proxAddr, timeout)
	if err != nil {
		return nil, err
	}

	server := &ProxyServer{
		addr: servAddr,
		prox: prox,
		log:  log,
	}

	switch servAddr.Scheme {
	case addr.HTTP:
		server.handleRequest = http.HandleRequest
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %q", servAddr.Scheme)
	}

	return server, nil
}

type ProxyServer struct {
	addr          *addr.Addr
	prox          *cli.ProxyClient
	log           *log.Logger
	handleRequest func(net.Conn, *cli.ProxyClient, *log.Logger)
	numReq        int
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
		log := s.log.WithAttr("id", strconv.Itoa(s.numReq))
		s.numReq++

		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("opening a client connection: %v", err)
			continue
		}

		go func() {
			defer func() {
				if err := cliConn.Close(); err != nil {
					log.Errorf("closing a client connection: %v", err)
				}
			}()
			s.handleRequest(cliConn, s.prox, log)
		}()
	}
}

func (s *ProxyServer) Addr() *addr.Addr {
	return s.addr
}
