package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

func main() {
	config := config.Load(os.Args)
	l := log.New(
		log.WithLevel(config.LogLevel),
	)

	client, err := proxy.NewClient(
		proxy.WithProxyAddr(&config.ProxyAddr),
	)
	if err != nil {
		l.Error("Failed to initialize a proxy client", err)
		return
	}
	l.Info("Using a proxy", "proxy_addr", &config.ProxyAddr)

	server, err := proxy.NewServer(
		proxy.WithListenAddr(&config.ServeAddr),
		proxy.WithServerDialer(client),
		proxy.WithServerLog(l),
	)
	if err != nil {
		l.Error("Failed to initialize a proxy server", err)
		return
	}

	server.Run(context.Background())
}
