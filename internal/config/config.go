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
	"github.com/cerfical/socks2http/internal/proxy/router"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var defaultConfig = Config{
	Server: ServerConfig{
		Proto: proxy.ProtoHTTP,
		Addr:  *addr.New("localhost", 8080),
	},

	Proxy: router.Proxy{
		Proto: proxy.ProtoDirect,
	},

	Log: LogConfig{
		Level: log.LevelVerbose,
	},
}

func Load(args []string) *Config {
	progName := getProgramName(args)

	flags := pflag.NewFlagSet(progName, pflag.ContinueOnError)
	flags.Usage = func() {
		fmt.Printf("Usage:\n")
		fmt.Printf("  %v [options]\n\n", progName)
		fmt.Printf("Options:\n")
		flags.PrintDefaults()
	}
	if err := parseFlags(flags, args); err != nil {
		printErrorAndExit(flags, err)
	}

	config, err := parseConfig(flags)
	if err != nil {
		printErrorAndExit(flags, err)
	}
	return config
}

func printErrorAndExit(f *pflag.FlagSet, err error) {
	fmt.Printf("Error: %v\n", err)
	f.Usage()
	os.Exit(1)
}

func parseConfig(f *pflag.FlagSet) (*Config, error) {
	v := viper.New()

	// Bind command-line flags to their corresponding values from config file
	configNames := []string{"server.addr", "server.proto", "proxy.addr", "proxy.proto", "log.level", "timeout"}
	for _, name := range configNames {
		kebabCasedName := strings.ReplaceAll(name, ".", "-")
		if err := v.BindPFlag(name, f.Lookup(kebabCasedName)); err != nil {
			panic(fmt.Errorf("bind flag: %w", err))
		}
	}

	v.SetConfigFile(f.Lookup("config-file").Value.String())
	if err := v.ReadInConfig(); err != nil {
		// Make the configuration file optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	}

	options := []viper.DecoderConfigOption{
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
		)),
	}

	config := defaultConfig
	if err := v.UnmarshalExact(&config, options...); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	return &config, nil
}

func parseFlags(f *pflag.FlagSet, args []string) error {
	// Flags shared with options from a configuration file
	c := defaultConfig
	f.Var(&textValue{&c.Server.Addr}, "server-addr", "address for proxy server to listen on")
	f.Var(&textValue{&c.Proxy.Addr}, "proxy-addr", "proxy address to connect via proxy client")
	f.Var(&textValue{&c.Server.Proto}, "server-proto", "proxy server protocol to use")
	f.Var(&textValue{&c.Proxy.Proto}, "proxy-proto", "proxy client protocol to use")
	f.Var(&textValue{&c.Log.Level}, "log-level", "severity level of logging messages")
	f.DurationVar(&c.Timeout, "timeout", c.Timeout, "``wait duration for I/O operations")

	help := f.Bool("help", false, "``display help message")
	f.String("config-file", "", "``configuration file")

	if err := f.Parse(args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *help {
		f.Usage()
		os.Exit(2)
	}
	return nil
}

func getProgramName(args []string) string {
	progPath := args[0]
	return strings.TrimSuffix(
		filepath.Base(progPath),
		filepath.Ext(progPath),
	)
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`

	Proxy  router.Proxy   `mapstructure:"proxy"`
	Routes []router.Route `mapstructure:"routes"`

	Log LogConfig `mapstructure:"log"`

	Timeout time.Duration `mapstructure:"timeout"`
}

type LogConfig struct {
	Level log.Level `mapstructure:"level"`
}

type ServerConfig struct {
	Proto proxy.Proto `mapstructure:"proto"`
	Addr  addr.Addr   `mapstructure:"addr"`
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
