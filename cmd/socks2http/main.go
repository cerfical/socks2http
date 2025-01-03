package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/internal/args"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox"
)

func main() {
	args := args.Parse(os.Args)
	log := log.New().WithLevel(args.LogLevel)

	proxy, err := prox.NewClient(&args.ProxyAddr)
	if err != nil {
		log.Fatalf("proxy init: %v", err)
	}

	log.Infof("using proxy %v", &args.ProxyAddr)
	log.Infof("starting server on %v", &args.ServerAddr)

	server, err := prox.NewServer(&args.ServerAddr, args.Timeout, proxy, log)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	if err := server.Run(context.Background()); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
