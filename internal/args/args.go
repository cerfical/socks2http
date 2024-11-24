package args

import (
	"cmp"
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

type Addr struct {
	Scheme   string
	Hostname string
	Port     string
}

func (a *Addr) Host() string {
	return a.Hostname + ":" + a.Port
}

var (
	Server   Addr
	Proxy    Addr
	Timeout  time.Duration
	LogLevel uint8
	UseProxy bool
)

func init() {
	serverAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defServerScheme, lookupPort(defServerScheme))}
	proxyAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defProxyScheme, lookupPort(defProxyScheme))}
	logLevel := flag.String("log-level", "error", "severity of logging messages")

	flag.Var(&serverAddr, "server-addr", "listen address for the server")
	flag.Var(&proxyAddr, "proxy-addr", "a proxy server to use")
	flag.BoolVar(&UseProxy, "use-proxy", false, "create a proxy chain")
	flag.DurationVar(&Timeout, "timeout", 0, "time to wait for a connection")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || serverAddr.isSet {
			util.FatalError("invalid command line options")
		}
		serverAddr.value = flag.Arg(0)
	}

	if !parseAddr(&Server, serverAddr.value, defServerScheme) {
		util.FatalError("invalid server address %q", serverAddr.value)
	}
	if !parseAddr(&Proxy, proxyAddr.value, defProxyScheme) {
		util.FatalError("invalid proxy address %q", proxyAddr.value)
	}
	UseProxy = UseProxy || proxyAddr.isSet

	switch *logLevel {
	case "fatal":
		LogLevel = LogFatal
	case "error":
		LogLevel = LogError
	case "info":
		LogLevel = LogInfo
	default:
		util.FatalError("invalid log level %v", *logLevel)
	}
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

func parseAddr(parsed *Addr, addr, scheme string) bool {
	matches := urlRegex.FindStringSubmatch(addr)
	if matches == nil {
		return false
	}

	*parsed = Addr{
		Scheme:   strings.ToLower(cmp.Or(matches[1], scheme)),
		Hostname: strings.ToLower(matches[2]),
		Port:     strings.ToLower(cmp.Or(matches[3], lookupPort(scheme))),
	}
	return true
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
