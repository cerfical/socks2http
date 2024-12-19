package args

import (
	"flag"
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
	var servAddr, proxAddr addr.Addr
	flag.TextVar(&servAddr, "serv", addr.New("http", "localhost", 8080), "listen address for the server")
	flag.TextVar(&proxAddr, "prox", addr.New("direct", "", 0), "a proxy server to use")
	timeout := flag.Duration("timeout", 0, "time to wait for a connection")

	var logLevel log.Level
	flag.TextVar(&logLevel, "log", log.Info, "severity `level` of logging messages")

	flag.Parse()

	return &Args{
		ServerAddr: &servAddr,
		ProxyAddr:  &proxAddr,
		LogLevel:   logLevel,
		Timeout:    *timeout,
	}, nil
}
