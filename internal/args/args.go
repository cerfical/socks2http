package args

import (
	"cmp"
	"flag"
	"fmt"
	"regexp"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"strings"
	"time"
)

const (
	defaultProxyScheme = "socks4"
	defaultServScheme  = "http"
	defaultHostname    = "localhost"
)

type Args struct {
	ServerAddr addr.Addr
	ProxyAddr  addr.Addr
	LogLevel   log.LogLevel
	Timeout    time.Duration
}

func Parse() (*Args, error) {
	servAddrFlag := flag.String("serv", "http", "listen address for the server")
	proxyAddrFlag := flag.String("proxy", "direct", "a proxy server to use")
	timeout := flag.Duration("timeout", 0, "time to wait for a connection")
	logLevelFlag := flag.String("log", "error", "severity of logging messages")
	flag.Parse()

	if narg := flag.NArg(); narg > 0 {
		if narg != 1 {
			return nil, fmt.Errorf("expected 1 positional argument, but got %v", narg)
		}
		*servAddrFlag = flag.Arg(0)
	}

	servAddr, err := parseAddr(*servAddrFlag, defaultServScheme)
	if err != nil {
		return nil, fmt.Errorf("server address: %w", err)
	}

	proxyAddr, err := parseAddr(*proxyAddrFlag, defaultProxyScheme)
	if err != nil {
		return nil, fmt.Errorf("proxy chain: %w", err)
	}

	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}

	return &Args{
		ServerAddr: *servAddr,
		ProxyAddr:  *proxyAddr,
		LogLevel:   logLevel,
		Timeout:    *timeout,
	}, nil
}

func parseAddr(addrStr, defaultScheme string) (*addr.Addr, error) {
	rawAddr, err := parseRawAddr(addrStr)
	if err != nil {
		return nil, err
	}

	rawAddr.scheme = cmp.Or(rawAddr.scheme, defaultScheme)
	rawAddr.hostname = cmp.Or(rawAddr.hostname, defaultHostname)

	portNum, err := portByScheme(rawAddr.scheme)
	if err != nil {
		return nil, err
	}

	// use the provided port if available, or the scheme's default port otherwise
	if rawAddr.port != "" {
		portNum, err = addr.ParsePort(rawAddr.port)
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

var addrRgx = regexp.MustCompile(`\A(?:(?<SCHEME>[^:]+):)?(?://)?(?<HOSTNAME>[^:]+)?(?::(?<PORT>[^:]+))?\z`)

func parseRawAddr(addrStr string) (addr rawAddr, err error) {
	matches := addrRgx.FindStringSubmatch(addrStr)
	if matches == nil {
		err = fmt.Errorf("invalid network address %q", addrStr)
	} else {
		addr = makeRawAddr(
			matches[addrRgx.SubexpIndex("SCHEME")],
			matches[addrRgx.SubexpIndex("HOSTNAME")],
			matches[addrRgx.SubexpIndex("PORT")],
		)
	}
	return
}

func makeRawAddr(scheme, hostname, port string) rawAddr {
	// normalize all address names to lowercase
	scheme = strings.ToLower(scheme)
	hostname = strings.ToLower(hostname)
	port = strings.ToLower(port)

	if hostname != "" {
		if scheme != "" {
			if port == "" {
				return makeRawAddr2(scheme, hostname)
			}
		} else if port != "" {
			return makeRawAddr2(hostname, port)
		} else {
			return makeRawAddr1(hostname)
		}
	}

	return rawAddr{
		scheme:   scheme,
		hostname: hostname,
		port:     port,
	}
}

func makeRawAddr1(str string) (raddr rawAddr) {
	switch {
	case isValidScheme(str):
		raddr.scheme = str
	case addr.IsValidPort(str):
		raddr.port = str
	default:
		raddr.hostname = str
	}
	return
}

func makeRawAddr2(str1, str2 string) (raddr rawAddr) {
	if isValidScheme(str1) {
		raddr.scheme = str1
		if addr.IsValidPort(str2) {
			raddr.port = str2
		} else {
			raddr.hostname = str2
		}
	} else {
		raddr.hostname = str1
		raddr.port = str2
	}
	return
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
	case "direct":
		return 0, nil
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
