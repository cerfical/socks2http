package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy/cli"
	"github.com/cerfical/socks2http/proxy/serv"
)

func main() {
	config := config.Load(os.Args)
	l := log.New(log.WithLevel(config.LogLevel))

	cli, err := cli.New(&config.ProxyAddr)
	if err != nil {
		l.Fatal("Failed to initialize a proxy client", err)
	}
	l.Info("Using a proxy", log.Fields{"addr": &config.ProxyAddr})

	serv, err := serv.New(&config.ServeAddr, config.Timeout, cli, l)
	if err != nil {
		l.Fatal("Failed to start up a server", err)
	}

	if err := serv.Serve(context.Background()); err != nil {
		l.Fatal("Server terminated abnormally", err)
	}
}
