package prox

import (
	"fmt"
	"net"
	"socks2http/internal/addr"
	"socks2http/internal/prox/socks"
	"time"
)

type Proxy struct {
	ProxyAddr addr.Addr
	Timeout   time.Duration
}

func (p *Proxy) Open(destAddr *addr.Addr) (net.Conn, error) {
	// if direct connection was requested, do not use a proxy
	if p.ProxyAddr.Scheme == addr.Direct {
		return net.DialTimeout("tcp", destAddr.Host(), p.Timeout)
	}

	// otherwise, establish a connection with an intermediate proxy
	proxyConn, err := net.DialTimeout("tcp", p.ProxyAddr.Host(), 0)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy %v: %w", p.ProxyAddr, err)
	}

	// and send a command for the proxy to connect to the destination server
	if err := connect(p.ProxyAddr.Scheme, proxyConn, destAddr, p.Timeout); err != nil {
		// ignore (?) the Close() error
		proxyConn.Close()
		return nil, err
	}

	return proxyConn, nil
}

func connect(proxyScheme addr.ProtoScheme, proxyConn net.Conn, destAddr *addr.Addr, timeout time.Duration) error {
	switch proxyScheme {
	case addr.SOCKS4:
		return socks.Connect(proxyConn, destAddr, timeout)
	default:
		return fmt.Errorf("unsupported client protocol scheme %q", proxyScheme)
	}
}

func Direct(timeout time.Duration) Proxy {
	return Proxy{
		ProxyAddr: addr.Addr{
			Scheme: addr.Direct,
		},
		Timeout: timeout,
	}
}
