package args

import (
	"cmp"
	"flag"
	"fmt"
	"net/url"
	"socks2http/util"
	"time"
)

const (
	defaultSOCKSProxyPort = "1080"
	defaultProxyPort      = defaultSOCKSProxyPort
	defaultProxyProto     = "socks4"

	defaultHTTPServerPort = "8080"
	defaultServerPort     = defaultHTTPServerPort
	defaultServerProto    = "http"
)

func Parse() *Args {
	args := Args{}

	serverAddr := stringArg{value: fmt.Sprintf("%v://localhost:%v", defaultServerProto, defaultServerPort)}
	proxyAddr := stringArg{value: fmt.Sprintf("%v://localhost:%v", defaultProxyProto, defaultProxyPort)}
	useProxy := flag.Bool("use-proxy", false, "create a proxy chain")
	flag.DurationVar(&args.Timeout, "timeout", 0, "time to wait for a connection")
	flag.Var(&serverAddr, "server-addr", "listen address for the server")
	flag.Var(&proxyAddr, "proxy-addr", "a proxy server to use")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 || serverAddr.isSet {
			util.FatalError("too many command line options")
		}
		serverAddr.value = flag.Arg(0)
	}

	server, err := newAddr(serverAddr.value, defaultServerProto, defaultServerPort)
	if err != nil {
		util.FatalError("invalid server address: %v", err)
	}
	args.Server = server

	if proxyAddr.isSet || *useProxy {
		args.Proxy, err = newAddr(proxyAddr.value, defaultProxyProto, defaultProxyPort)
		if err != nil {
			util.FatalError("invalid proxy server address: %v", err)
		}
	}
	return &args
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

type Args struct {
	Server  Addr
	Proxy   Addr
	Timeout time.Duration
}

type Addr struct {
	Host  string
	Proto string
}

type stringArg struct {
	isSet bool
	value string
}

func (a *stringArg) String() string {
	return a.value
}

func (a *stringArg) Set(val string) error {
	a.value = val
	a.isSet = true
	return nil
}
