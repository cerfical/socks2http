package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/args"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

func main() {
	args := args.Parse(os.Args)
	log := log.New().WithLevel(args.LogLevel)

	proxcli, err := proxy.NewClient(&args.ProxyAddr)
	if err != nil {
		log.Fatalf("proxy init: %v", err)
	}

	log.Infof("using proxy %v", &args.ProxyAddr)
	log.Infof("starting server on %v", &args.ServerAddr)

	proxserv, err := proxy.NewServer(&args.ServerAddr, args.Timeout, proxcli, log)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	if err := proxserv.Run(context.Background()); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
