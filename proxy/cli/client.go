package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
)

type Client struct {
	addr    *addr.Addr
	connect func(*bufio.Reader, io.Writer, *addr.Addr) error
}

func (c *Client) Open(ctx context.Context, destAddr *addr.Addr) (net.Conn, error) {
	// If direct connection was requested, do not use a proxy
	if c.addr.Scheme == addr.Direct {
		servConn, err := connect(ctx, destAddr)
		if err != nil {
			return nil, fmt.Errorf("connect to destination server: %w", err)
		}
		return servConn, nil
	}

	proxyConn, err := c.connectProxy(ctx, destAddr)
	if err != nil {
		return nil, err
	}
	return proxyConn, nil
}

func (c *Client) connectProxy(ctx context.Context, destAddr *addr.Addr) (net.Conn, error) {
	proxyConn, err := connect(ctx, c.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := proxyConn.SetDeadline(deadline); err != nil {
			_ = proxyConn.Close()
			return nil, err
		}
	}

	if err := c.connect(bufio.NewReader(proxyConn), proxyConn, destAddr); err != nil {
		_ = proxyConn.Close()
		return nil, fmt.Errorf("connect proxy to destination server: %w", err)
	}

	return proxyConn, nil
}

func (c *Client) Proto() string {
	return c.addr.Scheme
}

func New(proxyAddr *addr.Addr) (*Client, error) {
	c := &Client{addr: proxyAddr}
	switch c.addr.Scheme {
	case addr.Direct:
		c.connect = nil
	case addr.SOCKS4:
		c.connect = socksConnect
	case addr.HTTP:
		c.connect = httpConnect
	default:
		return nil, fmt.Errorf("unsupported client protocol scheme %v", c.addr.Scheme)
	}
	return c, nil
}

func connect(ctx context.Context, addr *addr.Addr) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr.Host())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func socksConnect(r *bufio.Reader, w io.Writer, destAddr *addr.Addr) error {
	if err := socks.WriteConnect(w, destAddr); err != nil {
		return err
	}
	return socks.ReadReply(r)
}

func httpConnect(r *bufio.Reader, w io.Writer, destAddr *addr.Addr) error {
	// with plain HTTP no preliminary connection is needed
	if destAddr.Scheme == addr.HTTP {
		return nil
	}

	// send HTTP CONNECT request
	connReq := http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: destAddr.Host()},
	}
	if err := connReq.Write(w); err != nil {
		return err
	}

	resp, err := http.ReadResponse(r, &connReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// discard the response body
	if _, err := io.ReadAll(resp.Body); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("%v %v", code, msg)
	}
	return nil
}
