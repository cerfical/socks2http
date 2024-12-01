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
			return nil, fmt.Errorf("server address: overriding %q with %q", servAddrFlag.value, flag.Arg(0))
		}
		servAddrFlag.value = flag.Arg(0)
	}

	servAddr, err := parseAddr(servAddrFlag.value, defaultServScheme)
	if err != nil {
		return nil, fmt.Errorf("server address: %w", err)
	}

	var proxyAddr *addr.Addr
	if *useProxy || proxyAddrFlag.isSet {
		proxyAddr, err = parseAddr(proxyAddrFlag.value, defaultProxyScheme)
		if err != nil {
			return nil, fmt.Errorf("proxy chain: %w", err)
		}
	}

	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
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
	rawAddr, err := parseRawAddr(addrStr)
	if err != nil {
		return nil, err
	}
	validateAddr(&rawAddr)

	// canonicalize domain and scheme names to lowercase
	rawAddr.scheme = strings.ToLower(cmp.Or(rawAddr.scheme, defaultScheme))
	rawAddr.hostname = strings.ToLower(cmp.Or(rawAddr.hostname, defaultHostname))

	portNum, err := portByScheme(rawAddr.scheme)
	if err != nil {
		return nil, err
	}

	// use the provided port if available, or the scheme's default port otherwise
	if rawAddr.port != "" {
		portNum, err = parsePort(rawAddr.port)
		if err != nil {
			return nil, fmt.Errorf("port number %q: %w", rawAddr.port, err)
		}
	}

	return &addr.Addr{
		Scheme: rawAddr.scheme,
		Host: addr.Host{
			Hostname: rawAddr.hostname,
			Port:     portNum,
		}}, nil
}

type rawAddr struct {
	scheme   string
	hostname string
	port     string
}

func validateAddr(addr *rawAddr) {
	if addr.hostname != "" {
		if addr.scheme != "" {
			if addr.port == "" {
				validateAddr2(addr, addr.scheme, addr.hostname)
			}
		} else if addr.port != "" {
			validateAddr2(addr, addr.hostname, addr.port)
		} else {
			validateAddr1(addr, addr.hostname)
		}
	}
}

func validateAddr1(addr *rawAddr, str string) {
	switch {
	case isValidScheme(str):
		addr.scheme = str
	case isValidPort(str):
		addr.port = str
	default:
		addr.hostname = str
	}
}

func validateAddr2(addr *rawAddr, str1, str2 string) {
	if isValidScheme(str1) {
		addr.scheme = str1
		if isValidPort(str2) {
			addr.port = str2
		} else {
			addr.hostname = str2
		}
	} else {
		addr.hostname = str1
		addr.port = str2
	}
}

var addrRgx = regexp.MustCompile(`\A(?:(?<SCHEME>[^:]+):)?(?://)?(?<HOSTNAME>[^:]+)?(?::(?<PORT>[^:]+))?\z`)

func parseRawAddr(addrStr string) (addr rawAddr, err error) {
	matches := addrRgx.FindStringSubmatch(addrStr)
	if matches == nil {
		err = fmt.Errorf("invalid network address %q", addrStr)
	} else {
		addr.scheme = matches[addrRgx.SubexpIndex("SCHEME")]
		addr.hostname = matches[addrRgx.SubexpIndex("HOSTNAME")]
		addr.port = matches[addrRgx.SubexpIndex("PORT")]
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
