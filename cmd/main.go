package main

import (
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/serv"
)

func main() {
	args, err := args.Parse()
	if err != nil {
		log.Fatal("command line: %v", err)
	}

	server, err := serv.New(args.ServerAddr, args.ProxyAddr, args.Timeout, log.New(args.LogLevel))
	if err != nil {
		log.Fatal("server init: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatal("server shutdown: %v", err)
	}
}
