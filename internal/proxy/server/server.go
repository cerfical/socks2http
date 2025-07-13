package server

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks"
)

func New(ops ...Option) *Server {
	defaults := []Option{
		WithDialer(proxy.DirectDialer),
		WithTunneler(proxy.DefaultTunneler),
		WithLogger(proxy.DiscardLogger),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}
	return &s
}

func WithDialer(d proxy.Dialer) Option {
	return func(s *Server) {
		s.dialer = d
	}
}

func WithTunneler(t proxy.Tunneler) Option {
	return func(s *Server) {
		s.tunneler = t
	}
}

func WithLogger(l proxy.Logger) Option {
	return func(s *Server) {
		s.log = l
	}
}

type Option func(*Server)

type Server struct {
	tunneler proxy.Tunneler
	dialer   proxy.Dialer

	log proxy.Logger
}

func (s *Server) ListenAndServe(ctx context.Context, serverURL *addr.URL) error {
	s.log.Info("Starting up a server")

	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", serverURL.Addr().String())
	if err != nil {
		return err
	}

	// Use an automatically assigned port if one was not specified
	listenPort := uint16(l.Addr().(*net.TCPAddr).Port)
	s.log.Info("Server is up", "server_url", addr.NewURL(serverURL.Proto, serverURL.Host, listenPort))

	return s.Serve(ctx, serverURL.Proto, l)
}

func (s *Server) Serve(ctx context.Context, p addr.Proto, l net.Listener) error {
	socksServ := SOCKSServer{
		Dialer:   s.dialer,
		Tunneler: s.tunneler,
		Log:      s.log,
	}

	httpServ := HTTPServer{
		Tunneler: s.tunneler,
		Dialer:   s.dialer,
		Log:      s.log,
	}

	switch p {
	case addr.ProtoSOCKS:
		return socksServ.ServeSOCKS(ctx, l)
	case addr.ProtoSOCKS4:
		socksServ.Version = socks.V4
		return socksServ.ServeSOCKS(ctx, l)
	case addr.ProtoSOCKS5:
		socksServ.Version = socks.V5
		return socksServ.ServeSOCKS(ctx, l)
	case addr.ProtoHTTP:
		return httpServ.ServeHTTP(ctx, l)
	default:
		_ = l.Close()
		return fmt.Errorf("unsupported protocol: %v", p)
	}
}
