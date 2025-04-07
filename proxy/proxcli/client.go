package proxcli

import (
	"context"
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

	if c.o.Addr.Scheme != addr.Direct {
		proxy, err := proxy.New(&c.o)
		if err != nil {
			return nil, err
		}
		c.proxy = proxy
	}

	return &c, nil
}

func WithProxyAddr(a *addr.Addr) Option {
	return func(c *Client) {
		c.o.Addr = *a
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(c *Client) {
		c.o.Dialer = d
	}
}

type Option func(*Client)

type Client struct {
	o proxy.Options

	proxy proxy.Proxy
}

func (c *Client) ProxyAddr() *addr.Addr {
	return &c.o.Addr
}

func (c *Client) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	// Connect to the server directly if no proxy is used
	if c.proxy == nil {
		return c.o.Dialer.Dial(ctx, h)
	}
	return c.proxy.Connect(ctx, h)
}
