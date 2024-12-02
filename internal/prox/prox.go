package prox

import (
	"fmt"
	"net"
	"socks2http/internal/addr"
	"socks2http/internal/prox/socks"
	"time"
)

type Proxy interface {
	Open(destHost addr.Host) (net.Conn, error)
}

func NewProxy(proxyAddr addr.Addr, timeout time.Duration) (Proxy, error) {
	switch proxyAddr.Scheme {
	case addr.SOCKS4:
		return socksProxy{
			host:    proxyAddr.Host(),
			timeout: timeout,
		}, nil
	case addr.Direct:
		return directProxy{
			timeout: timeout,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported client protocol scheme %q", proxyAddr.Scheme)
	}
}

type socksProxy struct {
	host    addr.Host
	timeout time.Duration
}

func (p socksProxy) Open(destHost addr.Host) (net.Conn, error) {
	return socks.ConnectTimeout(p.host, destHost, p.timeout)
}

type directProxy struct {
	timeout time.Duration
}

func (p directProxy) Open(destHost addr.Host) (net.Conn, error) {
	return net.DialTimeout("tcp", destHost.String(), p.timeout)
}
