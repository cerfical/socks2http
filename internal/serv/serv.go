package serv

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/cli"
	"github.com/cerfical/socks2http/internal/log"
)

func New(servAddr, proxAddr *addr.Addr, timeout time.Duration, log *log.Logger) (*ProxyServer, error) {
	prox, err := cli.New(proxAddr, timeout)
	if err != nil {
		return nil, err
	}

	requestReader, ok := newRequester(servAddr.Scheme)
	if !ok {
		return nil, fmt.Errorf("unsupported server protocol scheme %v", servAddr.Scheme)
	}

	return &ProxyServer{requestReader, servAddr, prox, log, 0}, nil
}

type ProxyServer struct {
	requester
	addr   *addr.Addr
	prox   *cli.ProxyClient
	log    *log.Logger
	numReq int
}

func (s *ProxyServer) Run() error {
	s.log.Infof("starting a server on %v", s.addr)
	if s.prox.IsDirect() {
		s.log.Infof("not using a proxy")
	} else {
		s.log.Infof("using a proxy %v", s.prox.Addr())
	}

	listener, err := net.Listen("tcp", s.addr.Host())
	if err != nil {
		return err
	}

	for {
		log := s.log.WithAttr("id", strconv.Itoa(s.numReq))
		s.numReq++

		closeWithMsgf := func(c io.Closer, fmt string, args ...any) {
			if err := c.Close(); err != nil {
				log.Errorf(fmt, args...)
			}
		}

		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("open a new client connection: %v", err)
			continue
		}

		// handle the request
		go func() {
			defer closeWithMsgf(cliConn, "close the client connection: %v", err)

			req, err := s.requester.request(cliConn)
			if err != nil {
				log.Errorf("parse the request: %v", err)
				return
			}
			defer closeWithMsgf(req, "clean up the request data: %v", err)

			servConn, err := s.prox.Open(req.destHost())
			if err != nil {
				log.Errorf("open a new server connection: %v", err)
				if err := req.writeReply(false); err != nil {
					log.Errorf("reject the request: %v", err)
				}
				return
			}
			defer closeWithMsgf(servConn, "close the server connection: %v", err)

			if err := req.writeReply(true); err != nil {
				log.Errorf("grant the request: %v", err)
				return
			}

			req.do(s.addr.Scheme, servConn, log)
		}()
	}
}
