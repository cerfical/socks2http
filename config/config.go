package config

import (
	"flag"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
)

var defaultServeAddr = addr.New(addr.HTTP, "localhost", 8080)
var defaultProxyAddr = addr.New(addr.Direct, "", 0)

// Config defines configurable application settings.
type Config struct {
	// ServeAddr is a listen address of the proxy server.
	ServeAddr addr.Addr

	// ProxyAddr is an address for the proxy client to connect to.
	ProxyAddr addr.Addr

	// Timeout specifies a timeout for I/O operations.
	Timeout time.Duration

	// LogLevel specifies verbosity level of log messages.
	LogLevel log.Level
}

// Load reads configuration options from command-line arguments.
func Load(args []string) *Config {
	var config Config

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.TextVar(&config.ServeAddr, "serve", defaultServeAddr, "listen `address` for the proxy server")
	flags.TextVar(&config.ProxyAddr, "proxy", defaultProxyAddr, "`address` of an optional intermediate proxy")
	flags.TextVar(&config.LogLevel, "log", log.LevelVerbose, "severity `level` of logging messages")
	flags.DurationVar(&config.Timeout, "timeout", 0, "wait time for I/O operations")

	// flag.ExitOnError: ignore errors
	_ = flags.Parse(args[1:])

	return &config
}
