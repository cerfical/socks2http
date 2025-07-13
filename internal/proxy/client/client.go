package client

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks"
)

func New(ops ...Option) *Client {
	defaults := []Option{
		WithDialer(proxy.DirectDialer),
	}

	var c Client
	for _, op := range slices.Concat(defaults, ops) {
		op(&c)
	}
	return &c
}

func WithProxyURL(u *addr.URL) Option {
	return func(c *Client) {
		c.proxyURL = *u
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(c *Client) {
		c.dialer = d
	}
}

type Option func(*Client)

type Client struct {
	proxyURL addr.URL
	dialer   proxy.Dialer
}

func (c *Client) Dial(ctx context.Context, dstAddr *addr.Addr) (net.Conn, error) {
	// Connect to destination directly if no proxy is used
	if c.proxyURL.IsZero() {
		return c.dialer.Dial(ctx, dstAddr)
	}

	// Connect to the proxy
	proxyConn, err := c.dialer.Dial(ctx, c.proxyURL.Addr())
	if err != nil {
		return nil, fmt.Errorf("dial proxy: %w", err)
	}

	// Connect the proxy to destination
	if err := c.connect(proxyConn, dstAddr); err != nil {
		proxyConn.Close()
		return nil, err
	}
	return proxyConn, nil
}

func (c *Client) connect(proxyConn net.Conn, dstAddr *addr.Addr) error {
	switch proto := c.proxyURL.Proto; proto {
	case addr.ProtoSOCKS4, addr.ProtoSOCKS4a:
		socksCli := SOCKSClient{socks.V4, proto == addr.ProtoSOCKS4}
		return socksCli.Connect(proxyConn, dstAddr)
	case addr.ProtoSOCKS5, addr.ProtoSOCKS5h:
		socksCli := SOCKSClient{socks.V5, proto == addr.ProtoSOCKS5}
		return socksCli.Connect(proxyConn, dstAddr)
	case addr.ProtoHTTP:
		httpCli := HTTPClient{}
		return httpCli.Connect(proxyConn, dstAddr)
	default:
		return fmt.Errorf("unsupported protocol: %v", proto)
	}
}
