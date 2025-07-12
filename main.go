package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cerfical/socks2http/internal/config"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/proxy/router"
	"github.com/cerfical/socks2http/internal/proxy/server"
)

func main() {
	config := config.Load(os.Args)
	log := log.New(log.WithLevel(config.Log.Level))

	log.Info("Using a proxy", "proxy_url", &config.Proxy)

	router := router.New(
		router.WithRoutes(config.Routes),
		router.WithDefaultRoute(&router.Route{
			Proxy: config.Proxy,
		}),
	)

	server := server.New(
		server.WithDialer(router),
		server.WithLogger(log),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := server.ListenAndServe(ctx, &config.Server); err != nil {
			log.Error("Server terminated abnormally", "error", err)
		}
	}()

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		// Wait for interrupts, and if one occurs, shut down the server
		log.Info("Shutting down the server")
		cancel()
	case <-done:
		// Server terminated abnormally
		return
	}

	select {
	case <-stop:
		// If another interrupt occurs, abort the shutdown and exit immediately
		log.Info("Shutdown aborted")
	case <-done:
		// The server shutdown was completed
		log.Info("Server is down")
	}
}
