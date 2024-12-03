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

	server := serv.Server{
		ProxyAddr: args.ProxyAddr,
		Timeout:   args.Timeout,
		Logger:    log.NewLogger(args.LogLevel),
	}

	if err := server.Run(args.ServerAddr); err != nil {
		log.Fatal("server shutdown: %v", err)
	}
}
