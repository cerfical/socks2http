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

	server := serv.NewServer()
	server.SetLogger(logger)
	server.SetTimeout(args.Timeout)
	server.SetProxy(args.ProxyAddr)

	if err := server.Run(args.ServerAddr); err != nil {
		logger.Fatal("server shutdown: %v", err)
	}
}
