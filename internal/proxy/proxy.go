package proxy

import (
	"fmt"
	"net"
	"net/url"
	"socks2http/internal/socks"
	"socks2http/internal/util"
	"strconv"
	"time"
)

type Proxy interface {
	open(destAddr string) (net.Conn, error)
}

func OpenURL(proxy Proxy, destServer *url.URL) (net.Conn, error) {
	port := destServer.Port()
	if port == "" {
		portNum, err := net.LookupPort("tcp", destServer.Scheme)
		if err != nil {
			return nil, fmt.Errorf("invalid destination server address %v: %w", destServer, err)
		}
		port = strconv.Itoa(portNum)
	}
	return proxy.open(destServer.Hostname() + ":" + port)
}

func NewProxy(proxyHost, proxyProto string, timeout time.Duration) Proxy {
	if proxyHost != "" {
		switch proxyProto {
		case "socks4":
			return &socks4Proxy{
				proxyHost: proxyHost,
				timeout:   timeout,
			}
		default:
			util.FatalError("unsupported proxy protocol scheme: %v", proxyProto)
		}
	}
	return &directProxy{timeout: timeout}
}

type socks4Proxy struct {
	proxyHost string
	timeout   time.Duration
}

func (p *socks4Proxy) open(destAddr string) (net.Conn, error) {
	return socks.ConnectTimeout(p.proxyHost, destAddr, p.timeout)
}

type directProxy struct {
	timeout time.Duration
}

func (p *directProxy) open(destAddr string) (net.Conn, error) {
	return net.DialTimeout("tcp", destAddr, p.timeout)
}
