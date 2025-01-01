package main

import (
	"context"
	"os"

	"github.com/cerfical/socks2http/internal/args"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/serv"
)

func main() {
	args := args.Parse(os.Args)

	logger := log.New().WithLevel(args.LogLevel)
	server, err := serv.New(&args.ServerAddr, &args.ProxyAddr, args.Timeout, logger)
	if err != nil {
		logger.Fatalf("server init: %v", err)
	}

	if err := server.Run(context.Background()); err != nil {
		logger.Fatalf("server shutdown: %v", err)
	}
}
