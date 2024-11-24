package args

import (
	"cmp"
	"flag"
	"fmt"
	"net/url"
	"socks2http/internal/util"
	"time"
)

const (
	defSOCKSProxyPort = "1080"
	defProxyPort      = defSOCKSProxyPort
	defProxyProto     = "socks4"

	defHTTPServerPort = "8080"
	defServerPort     = defHTTPServerPort
	defServerProto    = "http"
)

const (
	LogFatal = iota
	LogError
	LogInfo
)

type Addr struct {
	Host  string
	Proto string
}

var (
	Server   Addr
	Proxy    Addr
	Timeout  time.Duration
	LogLevel uint8
	UseProxy bool
)

func init() {
	serverAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defServerProto, defServerPort)}
	proxyAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defProxyProto, defProxyPort)}
	logLevel := flag.String("log-level", "error", "severity of logging messages")

	flag.Var(&serverAddr, "server-addr", "listen address for the server")
	flag.Var(&proxyAddr, "proxy-addr", "a proxy server to use")
	flag.BoolVar(&UseProxy, "use-proxy", false, "create a proxy chain")
	flag.DurationVar(&Timeout, "timeout", 0, "time to wait for a connection")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || serverAddr.isSet {
			util.FatalError("too many command line options")
		}
		serverAddr.value = flag.Arg(0)
	}

	var err error
	if Server, err = newAddr(serverAddr.value, defServerProto, defServerPort); err != nil {
		util.FatalError("invalid server address %v: %v", serverAddr, err)
	}
	if Proxy, err = newAddr(proxyAddr.value, defProxyProto, defProxyPort); err != nil {
		util.FatalError("invalid proxy server address %v: %v", proxyAddr, err)
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

func newAddr(addr, defProto, defPort string) (Addr, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		Host:  url.Hostname() + ":" + cmp.Or(url.Port(), defPort),
		Proto: cmp.Or(url.Scheme, defProto),
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
