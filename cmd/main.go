package main

import (
	"github.com/cerfical/socks2http/internal/args"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/serv"
)

func main() {
	logger := log.New()

	args, err := args.Parse()
	if err != nil {
		logger.Fatalf("command line: %v", err)
	}

	logger = logger.With().
		Level(args.LogLevel).
		Logger()

	server, err := serv.New(args.ServerAddr, args.ProxyAddr, args.Timeout, logger)
	if err != nil {
		logger.Fatalf("server init: %v", err)
	}

	if err := server.Run(); err != nil {
		logger.Fatalf("server shutdown: %v", err)
	}
}
