package main

import (
	"context"
	"net"
	"os"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/proxy/cli"
)

func main() {
	config := config.Load(os.Args)
	l := log.New(
		log.WithLevel(config.LogLevel),
	)

	client, err := cli.New(&config.ProxyAddr)
	if err != nil {
		l.Error("Failed to initialize a proxy client", err)
		return
	}
	l.Info("Using a proxy", log.Fields{"proxy_addr": &config.ProxyAddr})

	server, err := proxy.NewServer(
		proxy.WithListenAddr(&config.ServeAddr),
		proxy.WithDialer(proxy.DialerFunc(func(ctx context.Context, host string) (net.Conn, error) {
			hostname, port, err := net.SplitHostPort(host)
			if err != nil {
				return nil, err
			}

			portNum, err := addr.ParsePort(port)
			if err != nil {
				return nil, err
			}

			return client.Open(ctx, addr.New("", hostname, portNum))
		})),
		proxy.WithServerLog(l),
	)
	if err != nil {
		l.Error("Server initialization failure", err)
		return
	}

	server.Run(context.Background())
}
