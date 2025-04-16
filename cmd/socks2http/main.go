package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/proxy/proxcli"
	"github.com/cerfical/socks2http/proxy/proxserv"
)

func main() {
	config := config.Load(os.Args)
	log := log.New(log.WithLevel(config.LogLevel))

	client, err := proxcli.New(
		proxcli.WithProxyAddr(&config.ProxyAddr),
	)
	if err != nil {
		log.Error("Failed to initialize a proxy client", err)
		return
	}
	log.Info("Using a proxy", "addr", &config.ProxyAddr)

	server, err := proxserv.New(
		proxserv.WithProto(config.ServeAddr.Scheme),
		proxserv.WithProxy(proxy.New(client)),
		proxserv.WithLog(log),
	)
	if err != nil {
		log.Error("Failed to initialize a server", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := server.ListenAndServe(ctx, &config.ServeAddr.Host); err != nil {
			log.Error("Server terminated abnormally", err)
		}
	}()

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupts, and if one occurs, shut down the server
	<-stop
	log.Info("Shutting down the server")
	cancel()

	select {
	case <-stop:
		// If another interrupt occurs, abort the shutdown and exit immediately
		log.Info("Shutdown aborted")
	case <-done:
		// The server shutdown was completed
		log.Info("Server is down")
	}
}
