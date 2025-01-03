package prox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/socks"
)

func NewClient(proxyAddr *addr.Addr) (*Client, error) {
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

type Client struct {
	addr    *addr.Addr
	connect func(*bufio.Reader, io.Writer, *addr.Addr) error
}

func (c *Client) Open(ctx context.Context, destAddr *addr.Addr) (net.Conn, error) {
	// if direct connection was requested, do not use a proxy
	if c.addr.Scheme == addr.Direct {
		d := net.Dialer{}
		servConn, err := d.DialContext(ctx, "tcp", destAddr.Host())
		if err != nil {
			return nil, fmt.Errorf("connecting to server: %w", err)
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
	d := net.Dialer{}
	proxyConn, err := d.DialContext(ctx, "tcp", c.addr.Host())
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := proxyConn.SetDeadline(deadline); err != nil {
			_ = proxyConn.Close()
			return nil, err
		}
	}

	if err := c.connect(bufio.NewReader(proxyConn), proxyConn, destAddr); err != nil {
		_ = proxyConn.Close()
		return nil, fmt.Errorf("connecting proxy to server: %w", err)
	}

	return proxyConn, nil
}
