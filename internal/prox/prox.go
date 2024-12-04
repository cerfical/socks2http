package prox

import (
	"cmp"
	"fmt"
	"net"
	"socks2http/internal/addr"
	"socks2http/internal/prox/socks"
	"time"
)

func NewProxy(proxyAddr *addr.Addr, timeout time.Duration) (*Proxy, error) {
	proxy := &Proxy{
		addr:    proxyAddr,
		timeout: timeout,
	}

	// make sure some defaults are set
	proxy.addr.Scheme = cmp.Or(proxy.addr.Scheme, addr.SOCKS4)
	proxy.addr.Port = cmp.Or(proxy.addr.Port, proxy.addr.Scheme.Port())

	switch proxy.addr.Scheme {
	case addr.Direct:
		proxy.connect = func(net.Conn, *addr.Addr) error {
			return nil
		}
	case addr.SOCKS4:
		proxy.connect = socks.Connect
	default:
		return nil, fmt.Errorf("unsupported client protocol scheme %q", proxy.addr.Scheme)
	}

	return proxy, nil
}

func Direct(timeout time.Duration) *Proxy {
	return &Proxy{
		addr: &addr.Addr{
			Scheme: addr.Direct,
		},
		timeout: timeout,
	}
}

type Proxy struct {
	addr    *addr.Addr
	timeout time.Duration
	connect func(net.Conn, *addr.Addr) error
}

func (p *Proxy) Open(destAddr *addr.Addr) (net.Conn, error) {
	// if direct connection was requested, do not use a proxy
	if p.addr.Scheme == addr.Direct {
		return net.DialTimeout("tcp", destAddr.Host(), p.timeout)
	}

	// otherwise, establish a connection with an intermediate proxy
	proxyConn, err := net.DialTimeout("tcp", p.addr.Host(), 0)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy %v: %w", p.addr, err)
	}

	// and send a command for the proxy to connect to the destination server
	if err := p.connect(proxyConn, destAddr); err != nil {
		// ignore (?) the Close() error
		proxyConn.Close()
		return nil, err
	}
	return proxyConn, nil
}

func (p *Proxy) Addr() *addr.Addr {
	return p.addr
}
