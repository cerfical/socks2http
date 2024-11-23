package proxy

import (
	"fmt"
	"net"
	"net/url"
	"socks2http/internal/args"
	"socks2http/internal/socks"
	"socks2http/internal/util"
	"strconv"
)

func Open(destServer *url.URL) (net.Conn, error) {
	port := destServer.Port()
	if port == "" {
		portNum, err := net.LookupPort("tcp", destServer.Scheme)
		if err != nil {
			return nil, fmt.Errorf("invalid destination server address %v: %w", destServer, err)
		}
		port = strconv.Itoa(portNum)
	}
	return proxyOpen(destServer.Hostname() + ":" + port)
}

var proxyOpen func(string) (net.Conn, error)

func init() {
	if args.UseProxy {
		switch args.Proxy.Proto {
		case "socks4":
			proxyOpen = func(destAddr string) (net.Conn, error) {
				return socks.ConnectTimeout(args.Proxy.Host, destAddr, args.Timeout)
			}
		default:
			util.FatalError("unsupported proxy protocol scheme: %v", args.Proxy.Proto)
		}
	} else {
		proxyOpen = func(destAddr string) (net.Conn, error) {
			return net.DialTimeout("tcp", destAddr, args.Timeout)
		}
	}
}
