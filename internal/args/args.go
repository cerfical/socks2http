package args

import (
	"cmp"
	"flag"
	"fmt"
	"regexp"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"strconv"
	"strings"
	"time"
)

const (
	defProxyScheme = "socks4"
	defServScheme  = "http"
)

type Args struct {
	ServerAddr *addr.Addr
	ProxyAddr  *addr.Addr
	LogLevel   log.LogLevel
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	defServPort, err := lookupPort(defServScheme)
	if err != nil {
		return nil, err
	}

	defProxyPort, err := lookupPort(defProxyScheme)
	if err != nil {
		return nil, err
	}

	servAddrFlag := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defServScheme, defServPort)}
	flag.Var(&servAddrFlag, "server-addr", "listen address for the server")

	proxyAddrFlag := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defProxyScheme, defProxyPort)}
	flag.Var(&proxyAddrFlag, "proxy-addr", "a proxy server to use")
	useProxy := flag.Bool("use-proxy", false, "create a proxy chain")

	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	logLevelFlag := flag.String("log-level", "error", "severity of logging messages")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || servAddrFlag.isSet {
			return nil, fmt.Errorf("expected 1 positional argument, but got %v", narg)
		}
		servAddrFlag.value = flag.Arg(0)
	}

	servAddr, err := parseAddr(servAddrFlag.value, defServScheme)
	if err != nil {
		return nil, err
	}

	var proxyAddr *addr.Addr
	if *useProxy || proxyAddrFlag.isSet {
		if proxyAddr, err = parseAddr(proxyAddrFlag.value, defProxyScheme); err != nil {
			return nil, err
		}
	}

	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		return nil, err
	}

	return &Args{
		ServerAddr: servAddr,
		ProxyAddr:  proxyAddr,
		LogLevel:   logLevel,
		Timeout:    *timeout,
	}, nil
}

type stringFlag struct {
	isSet bool
	value string
}

func (f *stringFlag) String() string {
	return f.value
}

func (f *stringFlag) Set(val string) error {
	f.value = val
	f.isSet = true
	return nil
}

var urlRegex = regexp.MustCompile(`\A(?:([a-zA-Z0-9]+):)?(?://)?([-_.a-zA-Z0-9]+)(?::([0-9]+))?\z`)

func parseAddr(addrStr, defScheme string) (*addr.Addr, error) {
	matches := urlRegex.FindStringSubmatch(addrStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid address %q", addrStr)
	}

	scheme := strings.ToLower(cmp.Or(matches[1], defScheme))
	port, err := parsePort(matches[3], scheme)
	if err != nil {
		return nil, err
	}

	return &addr.Addr{
		Scheme: scheme,
		Host: addr.Host{
			Hostname: strings.ToLower(matches[2]),
			Port:     port,
		},
	}, nil
}

func parsePort(portStr, defScheme string) (uint16, error) {
	if portStr != "" {
		portNum, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return 0, fmt.Errorf("parsing port number %q: %w", portStr, err)
		}
		return uint16(portNum), nil
	}
	return lookupPort(defScheme)
}

func lookupPort(scheme string) (uint16, error) {
	switch scheme {
	case "socks4":
		return 1080, nil
	case "http":
		return 8080, nil
	default:
		return 0, fmt.Errorf("unknown protocol scheme %q", scheme)
	}
}

func parseLogLevel(logLevel string) (log.LogLevel, error) {
	switch logLevel {
	case "fatal":
		return log.LogFatal, nil
	case "error":
		return log.LogError, nil
	case "info":
		return log.LogInfo, nil
	default:
		return 0, fmt.Errorf("unknown log level %q", logLevel)
	}
}
