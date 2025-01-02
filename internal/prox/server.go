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
	s.log.Infof("using a proxy %v", s.proxy.addr)

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
			defer closeWithMsg(cliConn, "close a client connection", log)

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
	req, err := s.readRequest(ctx, cliConn)
	if err != nil {
		// ignore end-of-input errors
		if !errors.Is(err, io.EOF) {
			log.Errorf("read an incoming request: %v", err)
		}
		return
	}
	defer closeWithMsg(req, "clean up request data", log)

	req.logAttrs(log).Infof("incoming request")

	// open a connection to the destination server
	servConn, err := s.proxy.Open(ctx, req.destAddr())
	if err != nil {
		log.Errorf("open a server connection: %v", err)
		if req.isConnect() {
			if err := req.writeReject(cliConn); err != nil {
				log.Errorf("reject a connect request: %v", err)
			}
		}
		return
	}
	defer closeWithMsg(servConn, "close a server connection", log)

	// perform forwarding of incoming requests
	if req.isConnect() {
		if err := req.writeGrant(cliConn); err != nil {
			log.Errorf("grant a connect request: %v", err)
			return
		}

		if err := tunnelRequest(cliConn, servConn); err != nil {
			log.Errorf("tunnel a request: %v", err)
		}
	} else {
		if err := s.forwardRequest(req, cliConn, servConn); err != nil {
			log.Errorf("forward a request: %v", err)
		}
	}
}

func closeWithMsg(c io.Closer, msg string, log *log.Logger) {
	if err := c.Close(); err != nil {
		log.Errorf("%v: %v", msg, err)
	}
}

func (s *Server) forwardRequest(req request, cliConn, servConn net.Conn) error {
	if s.addr.Scheme == s.proxy.addr.Scheme {
		if err := req.writeProxy(servConn); err != nil {
			return err
		}
	} else {
		if err := req.write(servConn); err != nil {
			return err
		}
	}

	_, err := io.Copy(cliConn, servConn)
	return err
}

func tunnelRequest(cliConn, servConn net.Conn) error {
	errChan := make(chan error)
	go transfer(servConn, cliConn, errChan)
	go transfer(cliConn, servConn, errChan)

	for err := range errChan {
		if err != nil {
			resetConn(cliConn)
			resetConn(servConn)
			return err
		}
	}

	return nil
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func resetConn(conn net.Conn) {
	_ = conn.(*net.TCPConn).SetLinger(0)
}

func (s *Server) readRequest(ctx context.Context, cliConn net.Conn) (request, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := cliConn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set I/O timeouts: %w", err)
		}
	}

	req, err := s.parseRequest(bufio.NewReader(cliConn))
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	return req, nil
}
