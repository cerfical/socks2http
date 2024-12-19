package args

import (
	"flag"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

// Args represents a set of accepted command line options.
type Args struct {
	// ServerAddr is a listen address of the proxy server.
	ServerAddr addr.Addr

	// ProxyAddr is an address for the proxy client to connect to.
	ProxyAddr addr.Addr

	// LogLevel specifies global logging level.
	LogLevel log.Level

	// Timeout specifies a timeout for IO operations.
	Timeout time.Duration
}

// Parse parses [Args] from a list of string arguments, exiting on failure.
func Parse(args []string) *Args {
	a := &Args{}

	flags := flag.NewFlagSet("options", flag.ExitOnError)
	flags.TextVar(&a.ServerAddr, "serv", &addr.Addr{Scheme: "http", Hostname: "localhost", Port: 8080}, "listen `address` for the server")
	flags.TextVar(&a.ProxyAddr, "prox", &addr.Addr{Scheme: "direct"}, "`address` of an additional intermediate proxy")
	flags.TextVar(&a.LogLevel, "log", log.Info, "severity `level` of logging messages")
	flags.DurationVar(&a.Timeout, "timeout", 0, "connection timeout duration")

	// ignore errors, due to flag.ExitOnError
	_ = flags.Parse(args)

	return a
}
