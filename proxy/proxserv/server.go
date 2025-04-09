package proxserv

import (
	"context"
	"errors"
	"io"
	"net"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

func New(ops ...Option) (*Server, error) {
	defaults := []Option{
		WithServeAddr(addr.New(addr.HTTP, "localhost", 8080)),
		WithDialer(proxy.Direct),
		WithLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	proxy, err := proxy.New(&proxy.Options{
		Proto:  s.serveAddr.Scheme,
		Dialer: s.dialer,
		Log:    s.log,
	})
	if err != nil {
		return nil, err
	}
	s.proxy = proxy

	return &s, nil
}

func WithServeAddr(a *addr.Addr) Option {
	return func(s *Server) {
		s.serveAddr = *a
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(s *Server) {
		s.dialer = d
	}
}

func WithLog(l *log.Logger) Option {
	return func(s *Server) {
		s.log = l
	}
}

type Option func(*Server)

type Server struct {
	dialer    proxy.Dialer
	serveAddr addr.Addr
	proxy     proxy.Proxy

	log *log.Logger
}

func (s *Server) Serve(ctx context.Context) error {
	s.log.Info("Starting up a server")

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", s.serveAddr.Host.String())
	if err != nil {
		return err
	}

	// Update the serving address with the automatically assigned port if one was not specified
	s.serveAddr.Host.Port = uint16(listener.Addr().(*net.TCPAddr).Port)
	s.log.Info("Server is up",
		"addr", &s.serveAddr,
	)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			s.log.Error("Failed to accept an incoming client connection", err)
			continue
		}

		go func() {
			defer clientConn.Close()

			if err := s.proxy.Serve(ctx, clientConn); err != nil {
				// Ignore unimportant errors
				if !errors.Is(err, io.EOF) {
					s.log.Error("Failed to serve a request", err)
				}
			}
		}()
	}
}
