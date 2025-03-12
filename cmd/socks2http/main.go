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
	l := log.New(log.WithLevel(config.LogLevel))

	proxcli, err := proxy.NewClient(&config.ProxyAddr)
	if err != nil {
		l.Fatal("proxy init", err)
	}

	l.Info("using proxy", log.Fields{"addr": &config.ProxyAddr})
	l.Info("starting server", log.Fields{"addr": &config.ServeAddr})

	proxserv, err := proxy.NewServer(&config.ServeAddr, config.Timeout, proxcli, l)
	if err != nil {
		l.Fatal("server init", err)
	}

	if err := proxserv.Run(context.Background()); err != nil {
		l.Fatal("server shutdown", err)
	}
}
