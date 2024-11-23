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

type Addr struct {
	Host  string
	Proto string
}

var (
	Server   Addr
	Proxy    Addr
	Timeout  time.Duration
	UseProxy bool
)

func init() {
	var err error
	serverAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defServerProto, defServerPort)}
	proxyAddr := stringFlag{value: fmt.Sprintf("%v://localhost:%v", defProxyProto, defProxyPort)}

	flag.BoolVar(&UseProxy, "use-proxy", false, "create a proxy chain")
	flag.DurationVar(&Timeout, "timeout", 0, "time to wait for a connection")
	flag.Var(&serverAddr, "server-addr", "listen address for the server")
	flag.Var(&proxyAddr, "proxy-addr", "a proxy server to use")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || serverAddr.isSet {
			util.FatalError("too many command line options")
		}
		serverAddr.value = flag.Arg(0)
	}

	if Server, err = newAddr(serverAddr.value, defServerProto, defServerPort); err != nil {
		util.FatalError("invalid server address: %v", err)
	}
	if Proxy, err = newAddr(proxyAddr.value, defProxyProto, defProxyPort); err != nil {
		util.FatalError("invalid proxy server address: %v", err)
	}
	UseProxy = UseProxy || proxyAddr.isSet
}

func newAddr(addr, defaultProto, defaultPort string) (Addr, error) {
	proxyURL, err := url.Parse(addr)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		Host:  proxyURL.Hostname() + ":" + cmp.Or(proxyURL.Port(), defaultPort),
		Proto: cmp.Or(proxyURL.Scheme, defaultProto),
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
