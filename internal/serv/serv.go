package serv

import (
	"context"
	"errors"
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
	prox, err := cli.New(proxAddr)
	if err != nil {
		return nil, err
	}

	req, ok := newRequester(servAddr.Scheme)
	if !ok {
		return nil, fmt.Errorf("unsupported server protocol scheme %v", servAddr.Scheme)
	}

	return &ProxyServer{req, servAddr, prox, timeout, log, 0}, nil
}

type ProxyServer struct {
	requester
	addr    *addr.Addr
	prox    *cli.ProxyClient
	timeout time.Duration
	log     *log.Logger
	numReq  int
}

func (s *ProxyServer) Run(ctx context.Context) error {
	s.log.Infof("starting a server on %v", s.addr)
	if s.prox.IsDirect() {
		s.log.Infof("not using a proxy")
	} else {
		s.log.Infof("using a proxy %v", s.prox.Addr())
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr.Host())
	if err != nil {
		return err
	}

	for {
		log := s.log.WithAttr("id", strconv.Itoa(s.numReq))
		s.numReq++

		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("open a client connection: %v", err)
			continue
		}

		go func() {
			ctx := ctx
			if s.timeout != 0 {
				c, cancel := context.WithTimeout(ctx, s.timeout)
				defer cancel()
				ctx = c
			}
			s.handleConn(ctx, cliConn, log)
		}()
	}
}

func (s *ProxyServer) handleConn(ctx context.Context, cliConn net.Conn, log *log.Logger) {
	closeWithMsg := func(c io.Closer, msg string) {
		if err := c.Close(); err != nil {
			log.Errorf("%v: %v", msg, err)
		}
	}
	defer closeWithMsg(cliConn, "close a client connection")

	if deadline, ok := ctx.Deadline(); ok {
		if err := cliConn.SetDeadline(deadline); err != nil {
			log.Errorf("set client I/O timeouts: %v", err)
		}
	}

	req, err := s.requester.request(cliConn)
	if err != nil {
		// ignore end-of-input errors
		if !errors.Is(err, io.EOF) {
			log.Errorf("parse an incoming request: %v", err)
		}
		return
	}
	defer closeWithMsg(req, "clean up request data")

	servConn, err := s.prox.Open(ctx, req.destAddr())
	if err != nil {
		log.Errorf("open a server connection: %v", err)
		if err := req.writeReply(false); err != nil {
			log.Errorf("reject a connect request: %v", err)
		}
		return
	}
	defer closeWithMsg(servConn, "close a server connection")

	if err := req.writeReply(true); err != nil {
		log.Errorf("grant a connect request: %v", err)
		return
	}

	req.perform(s.addr.Scheme, servConn, log)
}
