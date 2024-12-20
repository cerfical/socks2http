package main

import (
	"os"

	"github.com/cerfical/socks2http/internal/args"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox/serv"
)

func main() {
	args := args.Parse(os.Args)

	logger := log.New().WithLevel(args.LogLevel)
	server, err := serv.New(&args.ServerAddr, &args.ProxyAddr, args.Timeout, logger)
	if err != nil {
		logger.Fatalf("server init: %v", err)
	}

	if err := server.Run(); err != nil {
		logger.Fatalf("server shutdown: %v", err)
	}
}
