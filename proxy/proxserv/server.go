package proxserv

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"sync"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks"
)

func New(ops ...Option) (*Server, error) {
	defaults := []Option{
		WithProto(addr.HTTP),
		WithProxy(proxy.New(proxy.DirectDialer)),
		WithLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	switch s.proto {
	case addr.SOCKS4:
		s.serveConn = s.serveSOCKS
	case addr.HTTP:
		s.serveConn = s.serveHTTP
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %v", s.proto)
	}

	return &s, nil
}

func WithProxy(p proxy.Proxy) Option {
	return func(s *Server) {
		s.proxy = p
	}
}

func WithProto(proto string) Option {
	return func(s *Server) {
		s.proto = proto
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

	serveConn func(context.Context, net.Conn) error
	proxy     proxy.Proxy

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

				if err := s.serveConn(context.Background(), clientConn); err != nil {
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

func (s *Server) serveHTTP(ctx context.Context, clientConn net.Conn) error {
	req, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}
	defer req.Body.Close()

	s.log.Info("New HTTP request",
		"method", req.Method,
		"uri", req.RequestURI,
		"proto", req.Proto,
	)

	dstHost, err := hostFromHTTPRequest(req)
	if err != nil {
		return fmt.Errorf("lookup destination host: %w", err)
	}

	// Special case for HTTP CONNECT
	if req.Method == http.MethodConnect {
		done, err := s.proxy.OpenTunnel(ctx, clientConn, dstHost)
		if err != nil {
			writeHTTPStatus(http.StatusBadGateway, clientConn)
			return fmt.Errorf("open tunnel to %v: %w", dstHost, err)
		}
		writeHTTPStatus(http.StatusOK, clientConn)
		return <-done
	}

	// All other requests are forwarded to the destination server as is
	resp, err := s.proxy.ForwardHTTP(ctx, req, dstHost)
	if err != nil {
		writeHTTPStatus(http.StatusBadGateway, clientConn)
		return fmt.Errorf("forward HTTP to %v: %w", dstHost, err)
	}

	return resp.Write(clientConn)
}

func (s *Server) serveSOCKS(ctx context.Context, clientConn net.Conn) error {
	req, err := socks.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}

	// TODO: Make log.Logger automatically call String()
	s.log.Info("New SOCKS request",
		"version", req.Version.String(),
		"command", req.Command.String(),
		"host", req.DstAddr.String(),
	)

	done, err := s.proxy.OpenTunnel(ctx, clientConn, &req.DstAddr)
	if err != nil {
		writeSOCKSReply(socks.Rejected, clientConn)
		return fmt.Errorf("open tunnel to %v: %w", &req.DstAddr, err)
	}
	writeSOCKSReply(socks.Granted, clientConn)
	return <-done
}

func hostFromHTTPRequest(r *http.Request) (*addr.Host, error) {
	// For HTTP CONNECT requests, the host is in the Request URL
	if r.Method == http.MethodConnect {
		h, err := addr.ParseHost(r.URL.Host)
		if err != nil {
			return nil, fmt.Errorf("parse request URL: %w", err)
		}
		return h, nil
	}

	// For others, the request URL contains the full destination URL, including the scheme
	port := r.URL.Port()
	if port == "" {
		// If the URL contains no port, we can try to guess it by looking at the scheme
		portNum, err := net.LookupPort("tcp", r.URL.Scheme)
		if err != nil {
			return nil, fmt.Errorf("lookup port by scheme: %w", err)
		}
		return addr.NewHost(r.URL.Hostname(), uint16(portNum)), nil
	}

	// If the port is specified, we can use it directly
	portNum, err := addr.ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}

	return addr.NewHost(r.URL.Hostname(), portNum), nil
}

func writeHTTPStatus(status int, clientConn net.Conn) {
	r := http.Response{ProtoMajor: 1, ProtoMinor: 1}
	r.StatusCode = status

	r.Write(clientConn)
}

func writeSOCKSReply(s socks.Status, clientConn net.Conn) {
	reply := socks.NewReply(s, nil)
	reply.Write(clientConn)
}
