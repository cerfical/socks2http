package proxserv

import (
	"context"
	"net"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

func New(ctx context.Context, ops ...Option) (*Server, error) {
	defaults := []Option{
		WithListenAddr(addr.New(addr.HTTP, "localhost", 8080)),
		WithDialer(proxy.Direct),
		WithLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	proxy, err := proxy.New(&s.o)
	if err != nil {
		return nil, err
	}
	s.proxy = proxy

	s.o.Log.Info("Starting up a server")

	var lc net.ListenConfig
	if s.listener, err = lc.Listen(ctx, "tcp", s.o.Addr.Host.String()); err != nil {
		return nil, err
	}
	// Update the listen address with the allocated port, if zero port was specified
	s.o.Addr.Host.Port = uint16(s.listener.Addr().(*net.TCPAddr).Port)

	s.o.Log.Info("Server is up",
		"listen_addr", &s.o.Addr,
	)

	return &s, nil
}

func WithListenAddr(a *addr.Addr) Option {
	return func(s *Server) {
		s.o.Addr = *a
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(s *Server) {
		s.o.Dialer = d
	}
}

func WithLog(l *log.Logger) Option {
	return func(s *Server) {
		s.o.Log = l
	}
}

type Option func(*Server)

type Server struct {
	o proxy.Options

	listener net.Listener
	proxy    proxy.Proxy
}

func (s *Server) ListenAddr() *addr.Addr {
	return &s.o.Addr
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			s.o.Log.Error("Failed to accept an incoming client connection", err)
			continue
		}

		go func() {
			defer clientConn.Close()

			if err := s.proxy.Serve(ctx, clientConn); err != nil {
				s.o.Log.Error("Failed to serve a request", err)
			}
		}()
	}
}
