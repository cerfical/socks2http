package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/socks"
)

var defaultListenAddr = addr.New(addr.HTTP, "localhost", 8080)
var defaultServerDialer Dialer = DialerFunc(directDial)

func NewServer(ops ...ServerOption) (*Server, error) {
	defaults := []ServerOption{
		WithListenAddr(defaultListenAddr),
		WithServerDialer(defaultServerDialer),
		WithServerLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	switch scheme := s.listenAddr.Scheme; scheme {
	case addr.HTTP:
		s.serve = s.serveHTTP
	case addr.SOCKS4:
		s.serve = s.serveSOCKS4
	default:
		return nil, fmt.Errorf("unsupported protocol scheme %v", scheme)
	}

	return &s, nil
}

func WithListenAddr(a *addr.Addr) ServerOption {
	return func(s *Server) {
		s.listenAddr = *a
	}
}

func WithServerDialer(d Dialer) ServerOption {
	return func(s *Server) {
		s.dialer = d
	}
}

func WithServerLog(l *log.Logger) ServerOption {
	return func(s *Server) {
		s.log = l
	}
}

type ServerOption func(*Server)

type Dialer interface {
	Dial(ctx context.Context, host string) (net.Conn, error)
}

type DialerFunc func(context.Context, string) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, host string) (net.Conn, error) {
	return f(ctx, host)
}

type Server struct {
	listenAddr addr.Addr
	listener   net.Listener

	dialer Dialer
	serve  func(context.Context, net.Conn)

	log *log.Logger
}

func (s *Server) ListenAddr() *addr.Addr {
	return &s.listenAddr
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		s.log.Error("Server startup failure", err)
		return err
	}

	defer func() {
		if err := s.Stop(); err != nil {
			s.log.Error("Server shutdown failure", err)
		}
	}()

	if err := s.Serve(ctx); err != nil {
		s.log.Error("Server terminated abnormally", err)
		return err
	}

	return nil
}

func (s *Server) Start(ctx context.Context) error {
	var (
		lc  net.ListenConfig
		err error
	)

	s.log.Info("Starting up a server", nil)

	s.listener, err = lc.Listen(ctx, "tcp", s.listenAddr.Host.String())
	if err != nil {
		return err
	}
	// Update the listen address with the allocated port, if zero port was specified
	s.listenAddr.Host.Port = s.listener.Addr().(*net.TCPAddr).Port

	s.log.Info("Server is up", log.Fields{"listen_addr": &s.listenAddr})

	return nil
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			s.log.Error("Failed to accept an incoming client connection", err)
			continue
		}

		go func() {
			defer s.closeConn(clientConn)

			s.serve(ctx, clientConn)
		}()
	}
}

func (s *Server) serveSOCKS4(ctx context.Context, clientConn net.Conn) {
	req, err := socks.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		s.log.Error("SOCKS4 request parsing failure", err)
		return
	}

	s.logSOCKS4(req)

	serverHost := addr.NewHost(req.DestIP.String(), int(req.DestPort))
	serverConn, ok := s.openConn(ctx, serverHost)
	if !ok {
		s.replySOCKS4(socks.RequestRejectedOrFailed, clientConn)
		return
	}
	defer s.closeConn(serverConn)

	if s.replySOCKS4(socks.RequestGranted, clientConn) {
		s.tunnel(clientConn, serverConn)
	}
}

func (s *Server) logSOCKS4(r *socks.Request) {
	s.log.Info("Incoming SOCKS4 request", log.Fields{
		"command": "CONNECT",
		"host":    addr.NewHost(r.DestIP.String(), int(r.DestPort)),
	})
}

func (s *Server) replySOCKS4(code byte, clientConn net.Conn) bool {
	reply := socks.Reply{Code: code}
	if err := reply.Write(clientConn); err != nil {
		s.log.Error("Error sending a SOCKS4 reply", err)
		return false
	}
	return true
}

func (s *Server) serveHTTP(ctx context.Context, clientConn net.Conn) {
	req, ok := s.parseHTTP(clientConn)
	if !ok {
		return
	}
	defer req.Body.Close()

	s.logHTTP(req)

	serverHost, err := addr.ParseHost(req.Host)
	if err != nil {
		s.log.Error(fmt.Sprintf("HTTP request has a Host header with an invalid value %v", req.Host), err)
		return
	}

	serverConn, ok := s.openConn(ctx, serverHost)
	if !ok {
		s.replyHTTP(http.StatusBadGateway, clientConn)
		return
	}
	defer s.closeConn(serverConn)

	// Special case for HTTP CONNECT
	if req.Method == http.MethodConnect {
		if s.replyHTTP(http.StatusOK, clientConn) {
			s.tunnel(clientConn, serverConn)
		}
		return
	}

	// All other requests are forwarded to the destination server as is
	s.forwardHTTP(req, clientConn, serverConn)
}

func (s *Server) logHTTP(r *http.Request) {
	s.log.Info("Incoming HTTP request", log.Fields{
		"method": r.Method,
		"uri":    r.RequestURI,
		"proto":  r.Proto,
	})
}

func (s *Server) parseHTTP(clientConn net.Conn) (*http.Request, bool) {
	req, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			s.log.Error("HTTP request parsing failure", err)
		}
		return nil, false
	}
	return req, true
}

func (s *Server) replyHTTP(status int, clientConn net.Conn) bool {
	r := http.Response{ProtoMajor: 1, ProtoMinor: 1}
	r.StatusCode = status

	if err := r.Write(clientConn); err != nil {
		s.log.Error("Error sending an HTTP response", err)
		return false
	}

	return true
}

func (s *Server) forwardHTTP(r *http.Request, clientConn, serverConn net.Conn) {
	if err := r.Write(serverConn); err != nil {
		s.log.Error(fmt.Sprintf("Error forwarding an HTTP request to %v", r.Host), err)
		return
	}

	if _, err := io.Copy(clientConn, serverConn); err != nil {
		s.log.Error(fmt.Sprintf("Error forwarding an HTTP response from %v", r.Host), err)
		return
	}
}

func (s *Server) tunnel(clientConn, serverConn net.Conn) {
	errChan := make(chan error)
	go transfer(serverConn, clientConn, errChan)
	go transfer(clientConn, serverConn, errChan)

	for err := range errChan {
		if err != nil {
			s.log.Error("Proxy tunnel failure", err)
		}
	}
}

func (s *Server) openConn(ctx context.Context, h *addr.Host) (net.Conn, bool) {
	conn, err := s.dialer.Dial(ctx, h.String())
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to establish a connection with %v", h), err)
		return nil, false
	}
	return conn, true
}

func (s *Server) closeConn(conn net.Conn) {
	if err := conn.Close(); err != nil {
		s.log.Error("Failed to close a connection", err)
	}
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func directDial(ctx context.Context, host string) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
