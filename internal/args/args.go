package args

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"socks2http/internal/util"
	"strings"
	"time"
)

const (
	defProxyScheme  = "socks4"
	defServerScheme = "http"
)

const (
	LogFatal = iota
	LogError
	LogInfo
)

type Args struct {
	ServerAddr *util.Addr
	ProxyAddr  *util.Addr
	LogLevel   uint8
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	serverAddrFlag := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defServerScheme, lookupPort(defServerScheme))}
	flag.Var(&serverAddrFlag, "server-addr", "listen address for the server")

	proxyAddrFlag := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defProxyScheme, lookupPort(defProxyScheme))}
	flag.Var(&proxyAddrFlag, "proxy-addr", "a proxy server to use")
	useProxy := flag.Bool("use-proxy", false, "create a proxy chain")

	logLevelFlag := flag.String("log-level", "error", "severity of logging messages")
	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	flag.Parse()

	// check for positional arguments
	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || serverAddrFlag.isSet {
			return nil, errors.New("invalid number of positional arguments")
		}
		serverAddrFlag.value = flag.Arg(0)
	}

	serverAddr := parseAddr(serverAddrFlag.value, defServerScheme)
	if serverAddr == nil {
		return nil, fmt.Errorf("invalid server address %q", serverAddrFlag.value)
	}

	var proxyAddr *util.Addr
	if *useProxy || proxyAddrFlag.isSet {
		proxyAddr = parseAddr(proxyAddrFlag.value, defProxyScheme)
		if proxyAddr == nil {
			return nil, fmt.Errorf("invalid proxy address %q", proxyAddrFlag.value)
		}
	}

	logLevel, ok := parseLogLevel(*logLevelFlag)
	if !ok {
		return nil, fmt.Errorf("invalid log level %v", *logLevelFlag)
	}

	return &Args{
		ServerAddr: serverAddr,
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

func parseAddr(addr, scheme string) *util.Addr {
	matches := urlRegex.FindStringSubmatch(addr)
	if matches == nil {
		return nil
	}

	return &util.Addr{
		Scheme:   strings.ToLower(cmp.Or(matches[1], scheme)),
		Hostname: strings.ToLower(matches[2]),
		Port:     strings.ToLower(cmp.Or(matches[3], lookupPort(scheme))),
	}
}

func lookupPort(scheme string) string {
	switch scheme {
	case "socks4":
		return "1080"
	case "http":
		return "8080"
	default:
		panic(fmt.Sprintf("unknown protocol scheme %q", scheme))
	}
}

func parseLogLevel(logLevel string) (uint8, bool) {
	switch logLevel {
	case "fatal":
		return LogFatal, true
	case "error":
		return LogError, true
	case "info":
		return LogInfo, true
	default:
		return 0, false
	}
}
