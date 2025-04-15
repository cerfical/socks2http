package proxserv

import (
	"context"
	"errors"
	"io"
	"net"
	"slices"
	"sync"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

func New(ops ...Option) (*Server, error) {
	defaults := []Option{
		WithProto(addr.HTTP),
		WithDialer(proxy.Direct),
		WithLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	proxy, err := proxy.New(&proxy.Options{
		Proto:  s.proto,
		Dialer: s.dialer,
		Log:    s.log,
	})
	if err != nil {
		return nil, err
	}
	s.proxy = proxy

	return &s, nil
}

func WithProto(proto string) Option {
	return func(s *Server) {
		s.proto = proto
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
	proto string

	dialer proxy.Dialer
	proxy  proxy.Proxy

	log *log.Logger
}

func (s *Server) ListenAndServe(ctx context.Context, serveAddr *addr.Host) error {
	s.log.Info("Starting up a server")

	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", serveAddr.String())
	if err != nil {
		return err
	}

	// Use an automatically assigned port if one was not specified
	addr := addr.New(s.proto, serveAddr.Hostname, uint16(l.Addr().(*net.TCPAddr).Port))
	s.log.Info("Server is up",
		"addr", addr,
	)

	return s.Serve(ctx, l)
}

func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	var activeConns sync.WaitGroup
	go func() {
		for {
			activeConns.Add(1)

			clientConn, err := l.Accept()
			if err != nil {
				activeConns.Done()

				if errors.Is(err, net.ErrClosed) {
					break
				}
				s.log.Error("Failed to accept an incoming client connection", err)
				continue
			}

			go func() {
				defer func() {
					clientConn.Close()
					activeConns.Done()
				}()

				if err := s.proxy.Serve(context.Background(), clientConn); err != nil {
					// Ignore less important errors
					if !errors.Is(err, io.EOF) {
						s.log.Error("Failed to serve a request", err)
					}
				}
			}()
		}
	}()

	// Wait for server shutdown
	<-ctx.Done()
	err := l.Close()
	activeConns.Wait()
	return err
}
