package proxcli

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"slices"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks"
)

func New(ops ...Option) (*Client, error) {
	defaults := []Option{
		WithProxyAddr(addr.New(addr.HTTP, "localhost", 8080)),
		WithDialer(proxy.DirectDialer),
	}

	var c Client
	for _, op := range slices.Concat(defaults, ops) {
		op(&c)
	}

	switch proto := c.proxyAddr.Scheme; proto {
	case addr.SOCKS4:
		c.connect = makeSOCKSConnector(socks.V4)
	case addr.SOCKS4a:
		c.connect = makeSOCKSConnector(socks.V4a)
	case addr.HTTP:
		c.connect = connectHTTP
	case addr.Direct:
		c.connect = nil
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %v", proto)
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
	connect   connector
}

func (c *Client) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	// Connect to the server directly if no proxy is used
	if c.connect == nil {
		return c.dialer.Dial(ctx, h)
	}

	// Otherwise establish a connection to a proxy
	proxyHost := &c.proxyAddr.Host
	proxyConn, err := c.dialer.Dial(ctx, proxyHost)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy %v: %w", proxyHost, err)
	}

	// And connect the proxy to the destination server
	if err := c.connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("connecto to %v: %w", h, err)
	}

	return proxyConn, nil
}

type connector func(net.Conn, *addr.Host) error

func makeSOCKSConnector(v socks.Version) connector {
	return func(c net.Conn, h *addr.Host) error {
		return connectSOCKS(v, c, h)
	}
}

func connectSOCKS(v socks.Version, proxyConn net.Conn, h *addr.Host) error {
	dstHost := h
	if v == socks.V4 {
		// SOCKS4 doesn't support specifying a hostname as a destination address
		// We need to resolve it to an IP address first
		ip4, err := h.ResolveToIPv4()
		if err != nil {
			return fmt.Errorf("resolve destination: %w", err)
		}
		dstHost = addr.NewHost(ip4.String(), h.Port)
	}

	connReq := socks.NewRequest(v, socks.Connect, dstHost)
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("SOCKS CONNECT: %w", err)
	}

	reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
	if err != nil {
		return fmt.Errorf("SOCKS CONNECT reply: %w", err)
	}

	if reply.Code != socks.Granted {
		return fmt.Errorf("SOCKS CONNECT rejected: %v", reply)
	}
	return nil
}

func connectHTTP(proxyConn net.Conn, h *addr.Host) error {
	connReq, err := http.NewRequest(http.MethodConnect, "", nil)
	if err != nil {
		return err
	}
	connReq.Host = h.String()

	if err := connReq.WriteProxy(proxyConn); err != nil {
		return fmt.Errorf("HTTP CONNECT: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("HTTP CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("HTTP CONNECT rejected: %v %v", code, msg)
	}

	return nil
}
