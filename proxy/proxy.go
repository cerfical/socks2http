package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/socks"
)

var Direct Dialer = DialerFunc(directDial)

func New(ops *Options) (Proxy, error) {
	proxy := proxy{ops.Dialer, ops.Log}
	switch ops.Proto {
	case addr.SOCKS4:
		return ProxyFunc(proxy.ServeSOCKS), nil
	case addr.HTTP:
		return ProxyFunc(proxy.ServeHTTP), nil
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %v", ops.Proto)
	}
}

type Option func(*Options)

type Options struct {
	Proto  string
	Dialer Dialer

	Log *log.Logger
}

type Proxy interface {
	Serve(ctx context.Context, clientConn net.Conn) error
}

type ProxyFunc func(context.Context, net.Conn) error

func (f ProxyFunc) Serve(ctx context.Context, clientConn net.Conn) error {
	return f(ctx, clientConn)
}

type Dialer interface {
	Dial(context.Context, *addr.Host) (net.Conn, error)
}

type DialerFunc func(context.Context, *addr.Host) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	return f(ctx, h)
}

type proxy struct {
	dialer Dialer

	log *log.Logger
}

func (p *proxy) ServeHTTP(ctx context.Context, clientConn net.Conn) error {
	req, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}
	defer req.Body.Close()

	p.log.Info("New HTTP request",
		"method", req.Method,
		"uri", req.RequestURI,
		"proto", req.Proto,
	)

	serverHost, err := p.hostFromHTTPRequest(req)
	if err != nil {
		return fmt.Errorf("lookup destination host: %w", err)
	}

	serverConn, err := p.dialer.Dial(ctx, serverHost)
	if err != nil {
		p.writeHTTPReply(http.StatusBadGateway, clientConn)
		return fmt.Errorf("connect to %v: %w", serverHost, err)
	}
	defer serverConn.Close()

	// Special case for HTTP CONNECT
	if req.Method == http.MethodConnect {
		if err := p.writeHTTPReply(http.StatusOK, clientConn); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
		if err := tunnel(clientConn, serverConn); err != nil {
			return fmt.Errorf("proxy tunnel: %w", err)
		}
		return nil
	}

	// All other requests are forwarded to the destination server as is
	return p.forwardHTTPRequest(req, clientConn, serverConn)
}

func (p *proxy) hostFromHTTPRequest(r *http.Request) (*addr.Host, error) {
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

func (p *proxy) writeHTTPReply(status int, clientConn net.Conn) error {
	r := http.Response{ProtoMajor: 1, ProtoMinor: 1}
	r.StatusCode = status

	return r.Write(clientConn)
}

func (p *proxy) forwardHTTPRequest(r *http.Request, clientConn, serverConn net.Conn) error {
	if err := r.Write(serverConn); err != nil {
		return fmt.Errorf("forward request: %w", err)
	}

	if _, err := io.Copy(clientConn, serverConn); err != nil {
		return fmt.Errorf("forward response: %w", err)
	}

	return nil
}

func (p *proxy) ServeSOCKS(ctx context.Context, clientConn net.Conn) error {
	req, err := socks.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}

	// TODO: Make log.Logger automatically call String()
	p.log.Info("New SOCKS request",
		"version", req.Version.String(),
		"command", req.Command.String(),
		"host", req.Host.String(),
	)

	serverConn, err := p.dialer.Dial(ctx, &req.Host)
	if err != nil {
		p.writeSOCKSReply(socks.Rejected, clientConn)
		return fmt.Errorf("connect to %v: %w", &req.Host, err)
	}
	defer serverConn.Close()

	if err := p.writeSOCKSReply(socks.Granted, clientConn); err != nil {
		return fmt.Errorf("write reply: %w", err)
	}
	if err := tunnel(clientConn, serverConn); err != nil {
		return fmt.Errorf("proxy tunnel: %w", err)
	}

	return nil
}

func (p *proxy) writeSOCKSReply(s socks.Status, clientConn net.Conn) error {
	reply := socks.NewReply(s)
	return reply.Write(clientConn)
}

func tunnel(clientConn, serverConn net.Conn) error {
	errChan := make(chan error)
	go transfer(serverConn, clientConn, errChan)
	go transfer(clientConn, serverConn, errChan)

	// Wait for both transfers to finish
	var firstErr error
	for err := range errChan {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func directDial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", h.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
