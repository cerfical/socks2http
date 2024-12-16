package args

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

type Args struct {
	ServerAddr *addr.Addr
	ProxyAddr  *addr.Addr
	LogLevel   log.Level
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	servAddrFlag := flag.String("serv", "http", "listen address for the server")
	proxyAddrFlag := flag.String("prox", "direct", "a proxy server to use")
	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	logLevelFlag := flag.String("log", "info", "severity of logging messages")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 {
			return nil, fmt.Errorf("expected 1 positional argument, but got %v", narg)
		}
		*servAddrFlag = flag.Arg(0)
	}

	servAddr, err := addr.Parse(*servAddrFlag)
	if err != nil {
		return nil, fmt.Errorf("proxy server address %q: %w", *servAddrFlag, err)
	}

	proxyAddr, err := addr.Parse(*proxyAddrFlag)
	if err != nil {
		return nil, fmt.Errorf("proxy client address %q: %w", *proxyAddrFlag, err)
	}

	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}

	return &Args{
		ServerAddr: servAddr,
		ProxyAddr:  proxyAddr,
		LogLevel:   logLevel,
		Timeout:    *timeout,
	}, nil
}

func parseLogLevel(lvl string) (log.Level, error) {
	switch logLevel := strings.ToLower(lvl); logLevel {
	case "fatal":
		return log.FatalLevel, nil
	case "error":
		return log.ErrorLevel, nil
	case "info":
		return log.InfoLevel, nil
	default:
		return 0, fmt.Errorf("unknown log level %q", logLevel)
	}
}
