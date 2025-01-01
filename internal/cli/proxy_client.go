package cli

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/socks"
)

func New(proxyAddr *addr.Addr, timeout time.Duration) (*ProxyClient, error) {
	prox := &ProxyClient{
		addr:    proxyAddr,
		timeout: timeout,
	}

	switch prox.addr.Scheme {
	case addr.Direct:
		prox.connect = nil
	case addr.SOCKS4:
		prox.connect = connectWithMsg(socks.Connect, "socks connect")
	case addr.HTTP:
		prox.connect = connectWithMsg(httpConnect, "http connect")
	default:
		return nil, fmt.Errorf("unsupported client protocol scheme %v", prox.addr.Scheme)
	}
	return prox, nil
}

func connectWithMsg(connect connectFunc, msg string) connectFunc {
	return func(proxConn net.Conn, destAddr *addr.Addr) (err error) {
		if err := connect(proxConn, destAddr); err != nil {
			return fmt.Errorf("%v: %w", msg, err)
		}
		return nil
	}
}

func httpConnect(proxConn net.Conn, destAddr *addr.Addr) (err error) {
	// with plain HTTP no preliminary connection is needed
	if destAddr.Scheme == addr.HTTP {
		return nil
	}

	// send HTTP CONNECT request
	connReq := http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: destAddr.Host()},
	}
	if err := connReq.Write(proxConn); err != nil {
		return fmt.Errorf("write a connect request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxConn), &connReq)
	if err != nil {
		return fmt.Errorf("read a connect response: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("clean up response data: %w", closeErr)
		}
	}()

	// discard the response body
	if _, err := io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("read response data: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("connect request failed: %v %v", code, msg)
	}
	return nil
}

type ProxyClient struct {
	addr    *addr.Addr
	timeout time.Duration
	connect connectFunc
}

type connectFunc func(net.Conn, *addr.Addr) error

func (p *ProxyClient) Open(destAddr *addr.Addr) (net.Conn, error) {
	// if direct connection was requested, do not use a proxy
	if p.addr.Scheme == addr.Direct {
		conn, err := net.DialTimeout("tcp", destAddr.Host(), p.timeout)
		if err != nil {
			return nil, fmt.Errorf("connect to %v: %w", destAddr.Host(), err)
		}
		return conn, nil
	}

	// otherwise, establish a connection with an intermediate proxy
	proxConn, err := net.DialTimeout("tcp", p.addr.Host(), p.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to %v: %w", p.addr.Host(), err)
	}

	// and send a command for the proxy to connect to the destination server
	if err := p.connect(proxConn, destAddr); err != nil {
		_ = proxConn.Close()
		return nil, fmt.Errorf("connect to %v: %w", destAddr.Host(), err)
	}

	return proxConn, nil
}

func (p *ProxyClient) Addr() *addr.Addr {
	return p.addr
}

func (p *ProxyClient) IsDirect() bool {
	return p.Proto() == addr.Direct
}

func (p *ProxyClient) Proto() string {
	return p.addr.Scheme
}
