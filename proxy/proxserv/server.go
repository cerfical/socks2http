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
	"strings"
	"sync"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks4"
	"github.com/cerfical/socks2http/socks5"
)

func New(ops ...Option) (*Server, error) {
	defaults := []Option{
		WithServeProto(addr.HTTP),
		WithProxy(proxy.New(proxy.DirectDialer)),
		WithLog(log.Discard),
	}

	var s Server
	for _, op := range slices.Concat(defaults, ops) {
		op(&s)
	}

	switch s.proto {
	case addr.SOCKS:
		s.serveConn = s.socksServe
	case addr.SOCKS4:
		s.serveConn = s.socks4Serve
	case addr.SOCKS5:
		s.serveConn = s.socks5Serve
	case addr.HTTP:
		s.serveConn = s.httpServe
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

func WithServeProto(proto string) Option {
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

	serveConn func(context.Context, *bufio.Reader, net.Conn) error
	proxy     proxy.Proxy

	log *log.Logger
}

func (s *Server) ListenAndServe(ctx context.Context, serveAddr *addr.Addr) error {
	s.log.Info("Starting up a server")

	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", serveAddr.String())
	if err != nil {
		return err
	}

	// Use an automatically assigned port if one was not specified
	addr := addr.New(serveAddr.Host, uint16(l.Addr().(*net.TCPAddr).Port))
	s.log.Info("Server is up", "proto", s.proto, "addr", addr)

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

				err := s.serveConn(context.Background(), bufio.NewReader(clientConn), clientConn)
				if err != nil {
					if errors.Is(err, io.EOF) {
						// Ignore connections closed by client
						return
					}
					s.log.Error(fmt.Sprintf("%v proxy failure", strings.ToUpper(s.proto)), err)
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

func (s *Server) httpServe(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) (err error) {
	req, err := http.ReadRequest(clientRead)
	if err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	defer req.Body.Close()

	s.log.Info("HTTP request", "method", req.Method, "uri", req.RequestURI)

	dstAddr, err := hostFromHTTPRequest(req)
	if err != nil {
		return fmt.Errorf("decode destination address: %w", err)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
	}

	var proxyErr error
	if req.Method == http.MethodConnect {
		// Special case for HTTP CONNECT
		tunnelDone, err := s.proxy.OpenTunnel(ctx, clientConn, dstAddr)
		if err != nil {
			proxyErr = fmt.Errorf("open tunnel: %w", err)
			resp.StatusCode = http.StatusBadGateway
		} else {
			defer func() {
				if tunnelErr := <-tunnelDone; tunnelErr != nil && err == nil {
					err = fmt.Errorf("close tunnel: %w", tunnelErr)
				}
			}()
		}
	} else {
		// All other requests are forwarded to destination as is
		r, err := s.proxy.ForwardHTTP(ctx, req, dstAddr)
		if err != nil {
			proxyErr = fmt.Errorf("HTTP forward: %w", err)
			resp.StatusCode = http.StatusBadGateway
		} else {
			resp = r
		}
	}

	s.log.Info("HTTP response", "status", makeHTTPStatusString(resp.StatusCode))

	if err := resp.Write(clientConn); err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	return proxyErr
}

func (s *Server) socks5Serve(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) (err error) {
	greet, err := socks5.ReadGreeting(clientRead)
	if err != nil {
		return fmt.Errorf("decode greeting: %w", err)
	}
	s.log.Info("SOCKS5 greeting", "auth_methods", greet.AuthMethods)

	greetReply := socks5.GreetingReply{
		AuthMethod: selectSOCKS5AuthMethod(greet.AuthMethods),
	}
	s.log.Info("SOCKS5 greeting reply", "auth_method", greetReply.AuthMethod)

	if err := greetReply.Write(clientConn); err != nil {
		return fmt.Errorf("encode greeting reply: %w", err)
	}

	if greetReply.AuthMethod == socks5.AuthNotAcceptable {
		return errors.New("no acceptable auth method was selected")
	}

	req, err := socks5.ReadRequest(clientRead)
	if err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	s.log.Info("SOCKS5 request", "command", req.Command, "destination", &req.DstAddr)

	reply := socks5.Reply{Status: socks5.StatusOK}
	var proxyErr error

	switch req.Command {
	case socks5.CommandConnect:
		tunnelDone, err := s.proxy.OpenTunnel(ctx, clientConn, &req.DstAddr)
		if err != nil {
			proxyErr = fmt.Errorf("open tunnel: %w", err)
			reply.Status = socks5.StatusHostUnreachable
		} else {
			defer func() {
				if tunnelErr := <-tunnelDone; tunnelErr != nil && err == nil {
					err = fmt.Errorf("close tunnel: %w", tunnelErr)
				}
			}()
		}
	default:
		reply.Status = socks5.StatusCommandNotSupported
	}

	s.log.Info("SOCKS5 reply", "status", reply.Status)

	if err := reply.Write(clientConn); err != nil {
		return fmt.Errorf("encode reply: %w", err)
	}
	return proxyErr
}

func (s *Server) socks4Serve(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) (err error) {
	req, err := socks4.ReadRequest(clientRead)
	if err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	s.log.Info("SOCKS4 request", "command", req.Command, "destination", &req.DstAddr)

	reply := socks4.Reply{Status: socks4.StatusGranted}
	var proxyErr error

	switch req.Command {
	case socks4.CommandConnect:
		tunnelDone, err := s.proxy.OpenTunnel(ctx, clientConn, &req.DstAddr)
		if err != nil {
			proxyErr = fmt.Errorf("open tunnel: %w", err)
			reply.Status = socks4.StatusRejectedOrFailed
		} else {
			defer func() {
				if tunnelErr := <-tunnelDone; tunnelErr != nil && err == nil {
					err = fmt.Errorf("close tunnel: %w", tunnelErr)
				}
			}()
		}
	default:
		reply.Status = socks4.StatusRejectedOrFailed
	}

	s.log.Info("SOCKS4 reply", "status", reply.Status)

	if err := reply.Write(clientConn); err != nil {
		return fmt.Errorf("encode reply: %w", err)
	}
	return proxyErr
}

func (s *Server) socksServe(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) error {
	version, err := clientRead.Peek(1)
	if err != nil {
		return fmt.Errorf("decode version: %w", err)
	}

	switch version[0] {
	case socks4.VersionCode:
		return s.socks4Serve(ctx, clientRead, clientConn)
	case socks5.VersionCode:
		return s.socks5Serve(ctx, clientRead, clientConn)
	default:
		return fmt.Errorf("invalid version (%#02x)", version[0])
	}
}

func hostFromHTTPRequest(r *http.Request) (*addr.Addr, error) {
	// For HTTP CONNECT requests, the host is in the Request URL
	if r.Method == http.MethodConnect {
		h, err := addr.Parse(r.URL.Host)
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
		return addr.New(r.URL.Hostname(), uint16(portNum)), nil
	}

	// If the port is specified, we can use it directly
	portNum, err := addr.ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}

	return addr.New(r.URL.Hostname(), portNum), nil
}

func selectSOCKS5AuthMethod(methods []socks5.AuthMethod) socks5.AuthMethod {
	for _, m := range methods {
		if m == socks5.AuthNone {
			return m
		}
	}
	return socks5.AuthNotAcceptable
}

func makeHTTPStatusString(status int) string {
	return fmt.Sprintf("%v %v", status, http.StatusText(status))
}
