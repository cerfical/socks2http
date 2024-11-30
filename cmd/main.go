package main

import (
	"log"
	"socks2http/internal/args"
	"socks2http/internal/proxy"
	"socks2http/internal/serv"
)

func main() {
	args, err := args.Parse()
	if err != nil {
		log.Fatalf("command line parsing: %v", err)
	}

	proxy, err := proxy.NewProxy(args.ProxyAddr, args.Timeout)
	if err != nil {
		log.Fatalf("failed to create proxy chain: %v", err)
	}

	server, err := serv.NewServer(args.ServerAddr, proxy, args.Timeout)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	servErrs := server.Run(args.LogLevel)
	for err := range servErrs {
		log.Printf("%v", err)
	}
}
