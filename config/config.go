package config

import (
	"flag"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
)

var defaultServeAddr = addr.New("localhost", 8080)
var defaultProxyAddr = addr.New("", 0)
var defaultProxyProto = proxy.ProtoDirect
var defaultServerProto = proxy.ProtoHTTP

func Load(args []string) *Config {
	var config Config

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.TextVar(&config.Server.Addr, "serve-addr", defaultServeAddr, "`address` to listen to by proxy server")
	flags.TextVar(&config.Proxy.Addr, "proxy-addr", defaultProxyAddr, "proxy server `address` to connect via proxy client")
	flags.TextVar(&config.Proxy.Proto, "proxy-proto", defaultProxyProto, "proxy client `protocol` to use")
	flags.TextVar(&config.Server.Proto, "serve-proto", defaultServerProto, "proxy server `protocol` to use")
	flags.TextVar(&config.LogLevel, "log", log.LevelVerbose, "severity `level` of logging messages")
	flags.DurationVar(&config.Timeout, "timeout", 0, "wait time for I/O operations")

	// flag.ExitOnError: ignore errors
	_ = flags.Parse(args[1:])

	return &config
}

type Config struct {
	Proxy  ProxyConfig
	Server ServerConfig

	Timeout  time.Duration
	LogLevel log.Level
}

type ServerConfig struct {
	Proto proxy.Proto
	Addr  addr.Addr
}
type ProxyConfig struct {
	Proto proxy.Proto
	Addr  addr.Addr
}
