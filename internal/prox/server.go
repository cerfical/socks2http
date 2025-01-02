package prox

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

func NewServer(servAddr, proxAddr *addr.Addr, timeout time.Duration, log *log.Logger) (*Server, error) {
	prox, err := NewClient(proxAddr)
	if err != nil {
		return nil, err
	}

	h, ok := newHandler(servAddr.Scheme)
	if !ok {
		return nil, fmt.Errorf("unsupported server protocol scheme %v", servAddr.Scheme)
	}

	return &Server{h, servAddr, prox, timeout, log, 0}, nil
}

type Server struct {
	handler
	addr    *addr.Addr
	proxy   *Client
	timeout time.Duration
	log     *log.Logger
	numReq  int
}

func (s *Server) Run(ctx context.Context) error {
	s.log.Infof("starting a server on %v", s.addr)
	s.log.Infof("using proxy %v", s.proxy.addr)

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
			log.Errorf("opening a client connection: %v", err)
			continue
		}

		go func() {
			defer closeWithMsg(cliConn, "closing a client connection", log)

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

func (s *Server) handleConn(ctx context.Context, cliConn net.Conn, log *log.Logger) {
	// read and validate an incoming request
	req, err := s.parseRequest(ctx, cliConn)
	if err != nil {
		// ignore end-of-input errors
		if !errors.Is(err, io.EOF) {
			log.Errorf("request parsing: %v", err)
		}
		return
	}

	defer func() {
		if closeErr := req.Close(); closeErr != nil {
			log.Errorf("%v", closeErr)
		}
	}()

	req.logAttrs(log).Infof("incoming request")

	// open a connection to the destination server
	servConn, err := s.proxy.Open(ctx, req.destAddr())
	if err != nil {
		log.Errorf("opening a server connection: %v", err)
		if err := req.writeReject(cliConn); err != nil {
			log.Errorf("%v", err)
		}
		return
	}
	defer closeWithMsg(servConn, "closing a server connection", log)

	if err := req.writeGrant(cliConn); err != nil {
		log.Errorf("%v", err)
		return
	}

	if err := req.do(cliConn, servConn, s.proxy.addr.Scheme); err != nil {
		log.Errorf("%v", err)
	}
}

func closeWithMsg(c io.Closer, msg string, log *log.Logger) {
	if err := c.Close(); err != nil {
		log.Errorf("%v: %v", msg, err)
	}
}

func (s *Server) parseRequest(ctx context.Context, cliConn net.Conn) (request, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := cliConn.SetDeadline(deadline); err != nil {
			return nil, err
		}
	}

	req, err := s.readRequest(bufio.NewReader(cliConn))
	if err != nil {
		return nil, err
	}
	return req, nil
}
