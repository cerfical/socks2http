package main

import (
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/serv"
)

func main() {
	logger := log.NewLogger()

	args, err := args.Parse()
	if err != nil {
		logger.Fatal("command line: %v", err)
	}
	logger.SetLevel(args.LogLevel)

	server := serv.Server{
		ProxyAddr: args.ProxyAddr,
		Timeout:   args.Timeout,
		Logger:    logger,
	}

	if err := server.Run(args.ServerAddr); err != nil {
		logger.Fatal("server shutdown: %v", err)
	}
}
