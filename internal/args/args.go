package args

import (
	"flag"
	"fmt"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"time"
)

type Args struct {
	ServerAddr *addr.Addr
	ProxyAddr  *addr.Addr
	LogLevel   log.LogLevel
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	servAddrFlag := flag.String("serv", "http", "listen address for the server")
	proxyAddrFlag := flag.String("proxy", "direct", "a proxy server to use")
	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	logLevelFlag := flag.String("log", "error", "severity of logging messages")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 {
			return nil, fmt.Errorf("expected 1 positional argument, but got %v", narg)
		}
		*servAddrFlag = flag.Arg(0)
	}

	servAddr, err := addr.ParseAddr(*servAddrFlag)
	if err != nil {
		return nil, fmt.Errorf("server address: %w", err)
	}

	proxyAddr, err := addr.ParseAddr(*proxyAddrFlag)
	if err != nil {
		return nil, fmt.Errorf("proxy chain: %w", err)
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

func parseLogLevel(logLevel string) (res log.LogLevel, err error) {
	switch logLevel {
	case "fatal":
		res = log.LogFatal
	case "error":
		res = log.LogError
	case "info":
		res = log.LogInfo
	default:
		err = fmt.Errorf("unknown log level %q", logLevel)
	}
	return
}
