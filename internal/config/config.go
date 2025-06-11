package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/router"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	defServerURL  = addr.NewURL(addr.ProtoHTTP, "localhost", 80)
	defProxyProto = addr.ProtoHTTP
	defLogLevel   = log.LevelVerbose
)

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

	rawConfig, err := parseRawConfig(flags)
	if err != nil {
		printErrorAndExit(flags, err)
	}
	return rawConfig.ToConfig()
}

func printErrorAndExit(f *pflag.FlagSet, err error) {
	fmt.Printf("Error: %v\n\n", err)
	f.Usage()
	os.Exit(1)
}

func parseRawConfig(f *pflag.FlagSet) (*rawConfig, error) {
	v := viper.New()

	// Bind command-line flags to their corresponding values from config file
	configNames := []string{"server", "proxy", "log.level", "timeout"}
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
			return nil, fmt.Errorf("load configuration: %w", err)
		}
	}

	options := []viper.DecoderConfigOption{
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
		)),

		func(c *mapstructure.DecoderConfig) {
			c.IgnoreUntaggedFields = true
		},
	}

	var config rawConfig
	if err := v.UnmarshalExact(&config, options...); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}
	return &config, nil
}

func parseFlags(f *pflag.FlagSet, args []string) error {
	// Flags shared with options from a configuration file
	serverURL := proxyURLValue(*defServerURL)
	f.Var(&serverURL, "server", "``address for proxy server to listen on")
	f.Var(&proxyURLValue{}, "proxy", "``proxy URL to connect via proxy client")

	logLevel := logLevelValue(defLogLevel)
	f.Var(&logLevel, "log-level", "``severity level of logging messages")

	f.Duration("timeout", 0, "``wait duration for I/O operations")

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
	Server addr.URL

	Proxy  addr.URL
	Routes []router.Route

	Log struct {
		Level log.Level
	}

	Timeout time.Duration
}

type rawConfig struct {
	Server proxyURLValue `mapstructure:"server"`

	Proxy  proxyURLValue `mapstructure:"proxy"`
	Routes []struct {
		Hosts []string      `mapstructure:"hosts"`
		Proxy proxyURLValue `mapstructure:"proxy"`
	} `mapstructure:"routes"`

	Log struct {
		Level logLevelValue `mapstructure:"level"`
	} `mapstructure:"log"`

	Timeout time.Duration `mapstructure:"timeout"`
}

func (c *rawConfig) ToConfig() *Config {
	var config Config

	config.Server = addr.URL(c.Server)
	config.Proxy = addr.URL(c.Proxy)
	config.Log.Level = log.Level(c.Log.Level)
	config.Timeout = c.Timeout

	for _, r := range c.Routes {
		route := router.Route{
			Hosts: r.Hosts,
			Proxy: addr.URL(r.Proxy),
		}
		config.Routes = append(config.Routes, route)
	}

	return &config
}

type proxyURLValue addr.URL

func (v *proxyURLValue) Set(s string) error {
	u, err := addr.ParseURL(s, defProxyProto)
	if err != nil {
		return err
	}
	*v = proxyURLValue(*u)
	return nil
}

func (v *proxyURLValue) UnmarshalText(text []byte) error {
	return v.Set(string(text))
}

func (v *proxyURLValue) String() string {
	return (*addr.URL)(v).String()
}

func (v *proxyURLValue) Type() string {
	return ""
}

type logLevelValue log.Level

func (v *logLevelValue) Set(s string) error {
	return (*log.Level)(v).UnmarshalText([]byte(s))
}

func (v *logLevelValue) UnmarshalText(text []byte) error {
	return v.Set(string(text))
}

func (v *logLevelValue) String() string {
	return (*log.Level)(v).String()
}

func (v *logLevelValue) Type() string {
	return ""
}
