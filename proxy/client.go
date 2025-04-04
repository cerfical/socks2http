package proxy

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
)

func NewClient(ops ...ClientOption) (*Client, error) {
	defaults := []ClientOption{
		WithProxyAddr(defaultListenAddr),
		WithClientDialer(defaultServerDialer),
	}

	var c Client
	for _, op := range slices.Concat(defaults, ops) {
		op(&c)
	}

	switch scheme := c.proxyAddr.Scheme; scheme {
	case addr.Direct:
		// Nothing to do, as there is no actual proxy server involved
		c.connect = nil
	case addr.SOCKS4:
		c.connect = connectSOCKS4
	case addr.HTTP:
		c.connect = connectHTTP
	default:
		return nil, fmt.Errorf("unsupported protocol scheme %v", scheme)
	}

	return &c, nil
}

func WithProxyAddr(a *addr.Addr) ClientOption {
	return func(c *Client) {
		c.proxyAddr = *a
	}
}

func WithClientDialer(d Dialer) ClientOption {
	return func(c *Client) {
		c.dialer = d
	}
}

type ClientOption func(*Client)

type Client struct {
	proxyAddr addr.Addr
	dialer    Dialer

	connect func(net.Conn, *addr.Host) error
}

func (c *Client) ProxyAddr() *addr.Addr {
	return &c.proxyAddr
}

func (c *Client) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	// Connect to the server directly
	if c.connect == nil {
		return c.dialer.Dial(ctx, h)
	}

	proxyHost := &c.proxyAddr.Host
	proxyConn, err := c.dialer.Dial(ctx, proxyHost)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy %v: %w", proxyHost, err)
	}

	if err := c.connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("connect to server %v: %w", h, err)
	}

	return proxyConn, nil
}

func connectHTTP(proxyConn net.Conn, h *addr.Host) error {
	connReq, err := http.NewRequest(http.MethodConnect, fmt.Sprintf("http://%v", h), nil)
	if err != nil {
		return fmt.Errorf("make HTTP CONNECT request: %w", err)
	}

	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("write HTTP CONNECT request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("read HTTP CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("unexpected HTTP CONNECT response: %v %v", code, msg)
	}

	return nil
}

func connectSOCKS4(proxyConn net.Conn, h *addr.Host) error {
	connReq := socks.NewRequest(socks.V4, socks.Connect, h)
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("write SOCKS CONNECT request: %w", err)
	}

	reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
	if err != nil {
		return fmt.Errorf("read SOCKS CONNECT reply: %w", err)
	}

	if reply != socks.Granted {
		return fmt.Errorf("%v", reply)
	}
	return nil
}
