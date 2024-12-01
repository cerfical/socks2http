package main

import (
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/proxy"
	"socks2http/internal/serv"
)

func main() {
	logger := log.NewLogger()

	args, err := args.Parse()
	if err != nil {
		logger.Fatal("command line: %v", err)
	}
	logger.SetLevel(args.LogLevel)

	proxy, err := proxy.NewProxy(args.ProxyAddr, args.Timeout)
	if err != nil {
		logger.Fatal("proxy chaining: %v", err)
	}

	server, err := serv.NewServer(args.ServerAddr, proxy, logger)
	if err != nil {
		logger.Fatal("server startup: %v", err)
	}

	if err := server.Run(); err != nil {
		logger.Fatal("server shutdown: %v", err)
	}
}
