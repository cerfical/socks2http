package proxcli

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
)

func New(ops ...Option) (*Client, error) {
	defaults := []Option{
		WithProxyAddr(addr.New(addr.HTTP, "localhost", 8080)),
		WithDialer(proxy.Direct),
	}

	var c Client
	for _, op := range slices.Concat(defaults, ops) {
		op(&c)
	}

	if proto := c.proxyAddr.Scheme; proto != addr.Direct {
		connector, err := proxy.NewConnector(proto)
		if err != nil {
			return nil, err
		}
		c.connector = connector
	}

	return &c, nil
}

func WithProxyAddr(a *addr.Addr) Option {
	return func(c *Client) {
		c.proxyAddr = *a
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(c *Client) {
		c.dialer = d
	}
}

type Option func(*Client)

type Client struct {
	proxyAddr addr.Addr
	dialer    proxy.Dialer
	connector proxy.Connector
}

func (c *Client) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	// Connect to the server directly if no proxy is used
	if c.connector == nil {
		return c.dialer.Dial(ctx, h)
	}

	// Otherwise establish a connection to a proxy
	proxyHost := &c.proxyAddr.Host
	proxyConn, err := c.dialer.Dial(ctx, proxyHost)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy %v: %w", proxyHost, err)
	}

	// And connect the proxy to the destination server
	if err := c.connector.Connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("connecto to %v: %w", h, err)
	}

	return proxyConn, nil
}
