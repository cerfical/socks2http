package serv

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"socks2http/internal/addr"
	"socks2http/internal/proxy"
	"strconv"
)

type LogLevel uint8

const (
	LogFatal LogLevel = iota
	LogError
	LogInfo
)

type Server interface {
	Run(logLevel LogLevel) <-chan error
}

func NewServer(servAddr *addr.Addr, proxy proxy.Proxy) (Server, error) {
	switch servAddr.Scheme {
	case "http":
		return &httpServerRunner{
			host:  servAddr.Host,
			proxy: proxy,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %q", servAddr.Scheme)
	}
}

type httpServerRunner struct {
	host  addr.Host
	proxy proxy.Proxy
}

func (r *httpServerRunner) Run(logLevel LogLevel) <-chan error {
	server := &httpServer{
		proxy:    r.proxy,
		errChan:  make(chan error),
		logLevel: logLevel,
	}

	go func() {
		defer close(server.errChan)
		if err := http.ListenAndServe(r.host.String(), server); err != nil {
			server.errChan <- fmt.Errorf("server shutdown: %w", err)
		}
	}()
	return server.errChan
}

type httpServer struct {
	proxy    proxy.Proxy
	errChan  chan error
	logLevel LogLevel
}

func (s *httpServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	s.info(requestLine)

	clientConn, _, err := wr.(http.Hijacker).Hijack()
	if err != nil {
		s.error("opening a client connection: %w", err)
		return
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			s.error("closing client connection: %w", err)
		}
	}()

	s.handleRequest(clientConn, req)
}

func (s *httpServer) handleRequest(clientConn net.Conn, req *http.Request) {
	destHost, err := extractHost(req.URL)
	if err != nil {
		s.error("extracting destination host from %q: %w", req.URL, err)
		return
	}

	servConn, err := s.proxy.Open(destHost)
	if err != nil {
		s.error("opening proxy to %v: %w", destHost, err)
		return
	}
	defer func() {
		if err := servConn.Close(); err != nil {
			s.error("closing proxy connection: %w", err)
		}
	}()

	if req.Method == http.MethodConnect {
		for err := range tunnel(clientConn, servConn) {
			s.error("tunnel to %v: %w", destHost, err)
		}
	} else {
		if err := sendRequest(clientConn, servConn, req); err != nil {
			s.error("sending request to %v: %w", destHost, err)
		}
	}
}

func (s *httpServer) error(format string, v ...any) {
	if s.logLevel >= LogError {
		s.errChan <- fmt.Errorf(format, v...)
	}
}

func (s *httpServer) info(format string, v ...any) {
	if s.logLevel >= LogInfo {
		s.errChan <- fmt.Errorf(format, v...)
	}
}

func extractHost(url *url.URL) (addr.Host, error) {
	port, err := parsePort(url.Port(), url.Scheme)
	if err != nil {
		return addr.Host{}, err
	}
	return addr.Host{Hostname: url.Hostname(), Port: port}, nil
}

func parsePort(portStr, scheme string) (uint16, error) {
	if portStr == "" {
		port, err := net.LookupPort("tcp", scheme)
		if err != nil {
			return 0, err
		}
		return uint16(port), nil
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

func sendRequest(clientConn, servConn net.Conn, req *http.Request) error {
	if err := req.Write(servConn); err != nil {
		return err
	}
	_, err := io.Copy(clientConn, servConn)
	return err
}
