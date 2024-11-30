package main

import (
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/proxy"
	"socks2http/internal/serv"
)

func main() {
	args, err := args.Parse()
	if err != nil {
		log.Fatal("command line parsing: %v", err)
	}

	proxy, err := proxy.NewProxy(args.ProxyAddr, args.Timeout)
	if err != nil {
		log.Fatal("failed to create proxy chain: %v", err)
	}

	server, err := serv.NewServer(args.ServerAddr, proxy, args.Timeout)
	if err != nil {
		log.Fatal("failed to start the server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatal("server shutdown: %v", err)
	}
}
