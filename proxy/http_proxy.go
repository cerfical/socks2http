package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/addr"
)

type httpProxy struct {
	*Options
}

func (p *httpProxy) Serve(ctx context.Context, clientConn net.Conn) error {
	req, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}
	defer req.Body.Close()

	p.Log.Info("New HTTP request",
		"method", req.Method,
		"uri", req.RequestURI,
		"proto", req.Proto,
	)

	serverHost, err := addr.ParseHost(req.Host)
	if err != nil {
		return fmt.Errorf("parse destination host: %w", err)
	}

	serverConn, err := p.Dialer.Dial(ctx, serverHost)
	if err != nil {
		p.writeReply(http.StatusBadGateway, clientConn)
		return fmt.Errorf("connect to %v: %w", serverHost, err)
	}
	defer serverConn.Close()

	// Special case for HTTP CONNECT
	if req.Method == http.MethodConnect {
		if err := p.writeReply(http.StatusOK, clientConn); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
		if err := tunnel(clientConn, serverConn); err != nil {
			return fmt.Errorf("proxy tunnel: %w", err)
		}
		return nil
	}

	// All other requests are forwarded to the destination server as is
	return p.forward(req, clientConn, serverConn)
}

func (p *httpProxy) writeReply(status int, clientConn net.Conn) error {
	r := http.Response{ProtoMajor: 1, ProtoMinor: 1}
	r.StatusCode = status

	return r.Write(clientConn)
}

func (p *httpProxy) forward(r *http.Request, clientConn, serverConn net.Conn) error {
	if err := r.Write(serverConn); err != nil {
		return fmt.Errorf("forward request: %w", err)
	}

	if _, err := io.Copy(clientConn, serverConn); err != nil {
		return fmt.Errorf("forward response: %w", err)
	}

	return nil
}

func (p *httpProxy) Connect(ctx context.Context, h *addr.Host) (net.Conn, error) {
	proxyHost := &p.Addr.Host
	proxyConn, err := p.Dialer.Dial(ctx, proxyHost)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy %v: %w", proxyHost, err)
	}

	if err := p.connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("HTTP CONNECT: %w", err)
	}

	return proxyConn, nil
}

func (p *httpProxy) connect(proxyConn net.Conn, h *addr.Host) error {
	connReq, err := http.NewRequest(http.MethodConnect, "", nil)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	connReq.Host = h.String()

	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("unexpected response: %v %v", code, msg)
	}

	return nil
}
