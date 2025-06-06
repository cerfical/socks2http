package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
)

func New(ops ...Option) (*Client, error) {
	defaults := []Option{
		WithProxyAddr(addr.New("localhost", 8080)),
		WithProxyProto(proxy.ProtoHTTP),
		WithDialer(proxy.DirectDialer),
	}

	var c Client
	for _, op := range slices.Concat(defaults, ops) {
		op(&c)
	}

	switch proto := c.proxyProto; proto {
	case proxy.ProtoSOCKS4, proxy.ProtoSOCKS4a:
		c.connect = func(c net.Conn, h *addr.Addr) error {
			return socks4Connect(c, h, proto == proxy.ProtoSOCKS4)
		}
	case proxy.ProtoSOCKS5, proxy.ProtoSOCKS5h:
		c.connect = func(c net.Conn, h *addr.Addr) error {
			return socks5Connect(c, h, proto == proxy.ProtoSOCKS5)
		}
	case proxy.ProtoHTTP:
		c.connect = httpConnect
	case proxy.ProtoDirect:
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

func WithProxyProto(p proxy.Proto) Option {
	return func(c *Client) {
		c.proxyProto = p
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(c *Client) {
		c.dialer = d
	}
}

type Option func(*Client)

type Client struct {
	proxyAddr  addr.Addr
	proxyProto proxy.Proto

	dialer  proxy.Dialer
	connect func(net.Conn, *addr.Addr) error
}

func (c *Client) Dial(ctx context.Context, h *addr.Addr) (net.Conn, error) {
	// Connect to destination directly if no proxy is used
	if c.connect == nil {
		return c.dialer.Dial(ctx, h)
	}

	// Otherwise establish a connection to a proxy
	proxyConn, err := c.dialer.Dial(ctx, &c.proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("dial proxy %v: %w", &c.proxyAddr, err)
	}

	// And connect the proxy to destination
	if err := c.connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("connect %v: %w", c.proxyProto, err)
	}
	return proxyConn, nil
}

func socks5Connect(proxyConn net.Conn, dstAddr *addr.Addr, resolveLocally bool) error {
	if resolveLocally {
		ip, err := net.ResolveIPAddr("ip", dstAddr.Host)
		if err != nil {
			return fmt.Errorf("resolve destination: %w", err)
		}
		dstAddr = addr.New(ip.String(), dstAddr.Port)
	}

	proxyRead := bufio.NewReader(proxyConn)
	greet := socks5.Greeting{
		AuthMethods: []socks5.AuthMethod{socks5.AuthNone},
	}
	if err := greet.Write(proxyConn); err != nil {
		return fmt.Errorf("encode greeting: %w", err)
	}

	greetReply, err := socks5.ReadGreetingReply(proxyRead)
	if err != nil {
		return fmt.Errorf("decode greeting reply: %w", err)
	}

	switch greetReply.AuthMethod {
	case socks5.AuthNone:
		// No authentication required
	case socks5.AuthNotAcceptable:
		return errors.New("no acceptable auth method was selected")
	default:
		return fmt.Errorf("unsupported auth method: %v", greetReply.AuthMethod)
	}

	connReq := socks5.Request{
		Command: socks5.CommandConnect,
		DstAddr: *dstAddr,
	}
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	reply, err := socks5.ReadReply(proxyRead)
	if err != nil {
		return fmt.Errorf("decode reply: %w", err)
	}

	if reply.Status != socks5.StatusOK {
		return fmt.Errorf("connection rejected: %v", reply.Status)
	}
	return nil
}

func socks4Connect(proxyConn net.Conn, dstAddr *addr.Addr, resolveLocally bool) error {
	if resolveLocally {
		ip4, err := net.ResolveIPAddr("ip4", dstAddr.Host)
		if err != nil {
			return fmt.Errorf("resolve destination: %w", err)
		}
		dstAddr = addr.New(ip4.String(), dstAddr.Port)
	}

	connReq := socks4.Request{
		Command: socks4.CommandConnect,
		DstAddr: *dstAddr,
	}
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	reply, err := socks4.ReadReply(bufio.NewReader(proxyConn))
	if err != nil {
		return fmt.Errorf("decode reply: %w", err)
	}

	if reply.Status != socks4.StatusGranted {
		return fmt.Errorf("connection rejected: %v", reply.Status)
	}
	return nil
}

func httpConnect(proxyConn net.Conn, dstAddr *addr.Addr) error {
	connReq, err := http.NewRequest(http.MethodConnect, "", nil)
	if err != nil {
		return err
	}
	connReq.Host = dstAddr.String()

	if err := connReq.WriteProxy(proxyConn); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("connection rejected: %v %v", code, msg)
	}
	return nil
}
