package args

import (
	"cmp"
	"flag"
	"fmt"
	"regexp"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"strconv"
	"time"
)

const (
	defaultProxyScheme = "socks4"
	defaultServScheme  = "http"
	defaultHostname    = "localhost"

	defaultServ  = "http://localhost:8080"
	defaultProxy = "socks4://localhost:1080"
)

type Args struct {
	ServerAddr *addr.Addr
	ProxyAddr  *addr.Addr
	LogLevel   log.LogLevel
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	servAddrFlag := stringFlag{value: defaultServ}
	flag.Var(&servAddrFlag, "serv", "listen address for the server")

	proxyAddrFlag := stringFlag{value: defaultProxy}
	flag.Var(&proxyAddrFlag, "proxy", "a proxy server to use")
	useProxy := flag.Bool("use-proxy", false, "create a proxy chain")

	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	logLevelFlag := flag.String("log", "error", "severity of logging messages")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 {
			return nil, fmt.Errorf("expected 1 positional argument, but got %v", narg)
		}
		if servAddrFlag.isSet {
			return nil, fmt.Errorf("overriding serv flag %q with %q", servAddrFlag.value, flag.Arg(0))
		}
		servAddrFlag.value = flag.Arg(0)
	}

	servAddr, err := parseAddr(servAddrFlag.value, defaultServScheme)
	if err != nil {
		return nil, err
	}

	var proxyAddr *addr.Addr
	if *useProxy || proxyAddrFlag.isSet {
		proxyAddr, err = parseAddr(proxyAddrFlag.value, defaultProxyScheme)
		if err != nil {
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

func parseAddr(addrStr, defaultScheme string) (*addr.Addr, error) {
	scheme, hostname, port, err := splitAddr(addrStr)
	if err != nil {
		return nil, err
	}

	scheme, hostname, port = arrangeAddr(scheme, hostname, port)
	scheme = cmp.Or(scheme, defaultScheme)
	hostname = cmp.Or(hostname, defaultHostname)

	portNum, err := portByScheme(scheme)
	if err != nil {
		return nil, err
	}

	if port != "" {
		portNum, err = parsePort(port)
		if err != nil {
			return nil, fmt.Errorf("port number %q: %w", port, err)
		}
	}

	return &addr.Addr{
		Scheme: scheme,
		Host: addr.Host{
			Hostname: hostname,
			Port:     portNum,
		}}, nil
}

func arrangeAddr(scheme, hostname, port string) (string, string, string) {
	if hostname != "" {
		if scheme != "" {
			if port == "" {
				return arrangeAddr2(scheme, hostname)
			}
		} else if port != "" {
			return arrangeAddr2(hostname, port)
		} else {
			return arrangeAddrP1(hostname)
		}
	}
	return scheme, hostname, port
}

func arrangeAddrP1(str string) (string, string, string) {
	switch {
	case isValidScheme(str):
		return str, "", ""
	case isValidPort(str):
		return "", "", str
	default:
		return "", str, ""
	}
}

func arrangeAddr2(str1, str2 string) (string, string, string) {
	if isValidScheme(str1) {
		if isValidPort(str2) {
			return str1, "", str2
		}
		return str1, str2, ""
	}
	return "", str1, str2
}

var addrRgx = regexp.MustCompile(`\A(?:(?<SCHEME>[^:]+):)?(?://)?(?<HOSTNAME>[^:]+)?(?::(?<PORT>[^:]+))?\z`)

func splitAddr(addr string) (scheme string, hostname string, port string, err error) {
	matches := addrRgx.FindStringSubmatch(addr)
	if matches == nil {
		err = fmt.Errorf("invalid network address %q", addr)
	} else {
		scheme = matches[addrRgx.SubexpIndex("SCHEME")]
		hostname = matches[addrRgx.SubexpIndex("HOSTNAME")]
		port = matches[addrRgx.SubexpIndex("PORT")]
	}
	return
}

func isValidPort(port string) bool {
	_, err := parsePort(port)
	return err == nil
}

func parsePort(port string) (uint16, error) {
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("not a 16-bit unsigned integer: %w", err)
	}
	return uint16(portNum), nil
}

func isValidScheme(scheme string) bool {
	_, err := portByScheme(scheme)
	return err == nil
}

func portByScheme(scheme string) (uint16, error) {
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
