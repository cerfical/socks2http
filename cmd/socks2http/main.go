package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy/proxcli"
	"github.com/cerfical/socks2http/proxy/proxserv"
)

func main() {
	config := config.Load(os.Args)
	log := log.New(
		log.WithLevel(config.LogLevel),
	)

	client, err := proxcli.New(
		proxcli.WithProxyAddr(&config.ProxyAddr),
	)
	if err != nil {
		log.Error("Failed to initialize a proxy client", err)
		return
	}

	log.Info("Using a proxy",
		"proxy_addr", &config.ProxyAddr,
	)

	server, err := proxserv.New(context.Background(),
		proxserv.WithListenAddr(&config.ServeAddr),
		proxserv.WithDialer(client),
		proxserv.WithLog(log),
	)
	if err != nil {
		log.Error("Failed to start a proxy server", err)
		return
	}

	if err := server.Serve(context.Background()); err != nil {
		log.Error("Server terminated abnormally", err)
		return
	}
}
