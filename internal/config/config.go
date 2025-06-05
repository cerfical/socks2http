package config

import (
	"encoding"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/spf13/pflag"
)

func Load(args []string) *Config {
	config := Config{
		Proxy: ProxyConfig{
			Proto: proxy.ProtoDirect,
		},
		Server: ServerConfig{
			Proto: proxy.ProtoHTTP,
			Addr:  *addr.New("localhost", 8080),
		},
		Log: LogConfig{
			Level: log.LevelVerbose,
		},
	}

	progName := getProgramName(args)
	flags := pflag.NewFlagSet(progName, pflag.ContinueOnError)
	flags.Usage = func() {
		fmt.Printf("Usage:\n")
		fmt.Printf("  %v [options]\n\n", progName)
		fmt.Printf("Options:\n")
		flags.PrintDefaults()
	}

	flags.Var(&textValue{&config.Server.Addr}, "server-addr", "address for proxy server to listen on")
	flags.Var(&textValue{&config.Proxy.Addr}, "proxy-addr", "proxy address to connect via proxy client")
	flags.Var(&textValue{&config.Server.Proto}, "server-proto", "proxy server protocol to use")
	flags.Var(&textValue{&config.Proxy.Proto}, "proxy-proto", "proxy client protocol to use")
	flags.Var(&textValue{&config.Log.Level}, "log-level", "severity level of logging messages")
	flags.DurationVar(&config.Timeout, "timeout", 0, "``wait duration for I/O operations")

	help := flags.Bool("help", false, "``display help message")
	if err := flags.Parse(args[1:]); err != nil || *help {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		flags.Usage()
		os.Exit(2)
	}
	return &config
}

func getProgramName(args []string) string {
	progPath := args[0]
	return strings.TrimSuffix(
		filepath.Base(progPath),
		filepath.Ext(progPath),
	)
}

type Config struct {
	Proxy  ProxyConfig
	Server ServerConfig

	Log LogConfig

	Timeout time.Duration
}

type LogConfig struct {
	Level log.Level
}

type ServerConfig struct {
	Proto proxy.Proto
	Addr  addr.Addr
}

type ProxyConfig struct {
	Proto proxy.Proto
	Addr  addr.Addr
}

type textValue struct {
	val interface {
		encoding.TextMarshaler
		encoding.TextUnmarshaler
	}
}

func (v *textValue) Set(s string) error {
	return v.val.UnmarshalText([]byte(s))
}

func (v *textValue) String() string {
	return fmt.Sprintf("%v", v.val)
}

func (v *textValue) Type() string {
	return ""
}
