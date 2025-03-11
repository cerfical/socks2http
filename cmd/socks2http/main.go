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
	log := log.New().WithLevel(config.LogLevel)

	proxcli, err := proxy.NewClient(&config.ProxyAddr)
	if err != nil {
		log.Fatalf("proxy init: %v", err)
	}

	log.Infof("using proxy %v", &config.ProxyAddr)
	log.Infof("starting server on %v", &config.ServeAddr)

	proxserv, err := proxy.NewServer(&config.ServeAddr, config.Timeout, proxcli, log)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	if err := proxserv.Run(context.Background()); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
