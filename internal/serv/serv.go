package serv

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"socks2http/internal/addr"
	"socks2http/internal/proxy"
	"strconv"
	"sync"
	"time"
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

func NewServer(servAddr *addr.Addr, proxy proxy.Proxy, timeout time.Duration) (Server, error) {
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
		errors:   make(chan error),
		logLevel: logLevel,
	}

	go func() {
		defer close(server.errors)
		if err := http.ListenAndServe(r.host.String(), server); err != nil {
			server.errors <- fmt.Errorf("server shutdown: %w", err)
		}
	}()
	return server.errors
}

type httpServer struct {
	proxy    proxy.Proxy
	errors   chan error
	logLevel LogLevel
}

func (s *httpServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	s.info(requestLine)

	port, err := parsePort(req.URL.Port(), req.URL.Scheme)
	if err != nil {
		s.error("parsing destination server URL %q: %w", req.URL, err)
		return
	}
	destAddr := addr.Host{Hostname: req.URL.Hostname(), Port: port}

	proxyConn, err := s.proxy.Open(destAddr)
	if err != nil {
		s.error("connecting to destination server %v: %w", destAddr, err)
		return
	}
	defer func() {
		if err := proxyConn.Close(); err != nil {
			s.error("closing proxy connection: %w", err)
		}
	}()

	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		s.error("hijacking not supported")
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		s.error("failed: %w", err)
		return
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			s.error("closing client connection: %w", err)
		}
	}()

	if req.Method != http.MethodConnect {
		if err := sendRequest(clientConn, proxyConn, req); err != nil {
			s.error("sending request to %v: %w", destAddr, err)
		}
	} else {
		tunnelErrs, err := setupHttpTunnel(clientConn, proxyConn, req)
		if err != nil {
			s.error("HTTP tunnel setup to %v: %w", destAddr, err)
			return
		}

		for err := range tunnelErrs {
			s.error("HTTP tunnel to %v: %w", destAddr, err)
		}
	}
}

func (s *httpServer) error(format string, v ...any) {
	if s.logLevel >= LogError {
		s.errors <- fmt.Errorf(format, v...)
	}
}

func (s *httpServer) info(format string, v ...any) {
	if s.logLevel >= LogInfo {
		s.errors <- fmt.Errorf(format, v...)
	}
}

func parsePort(port, scheme string) (uint16, error) {
	if port != "" {
		port, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(port), nil
	} else {
		port, err := net.LookupPort("tcp", scheme)
		if err != nil {
			return 0, fmt.Errorf("invalid URL scheme %q: %w", scheme, err)
		}
		return uint16(port), nil
	}
}

func sendRequest(clientConn, proxyConn net.Conn, req *http.Request) error {
	if err := req.Write(proxyConn); err != nil {
		return err
	}
	_, err := io.Copy(clientConn, proxyConn)
	return err
}

func setupHttpTunnel(clientConn, proxyConn net.Conn, req *http.Request) (<-chan error, error) {
	okResp := http.Response{
		StatusCode: http.StatusOK,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
	}
	if err := okResp.Write(clientConn); err != nil {
		return nil, fmt.Errorf("replying to a CONNECT: %w", err)
	}

	errChan := make(chan error)
	go httpTunnel(clientConn, proxyConn, errChan)
	return errChan, nil
}

func httpTunnel(clientConn, proxyConn net.Conn, errChan chan<- error) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	defer func() {
		wg.Wait()
		close(errChan)
	}()

	transfer := func(dest, src net.Conn) {
		defer wg.Done()

		reportError := func(conn net.Conn, err error) {
			// use deadlines to preemptively terminate Read()/Write() calls and prevent goroutines being blocked indefinitely
			if errors.Is(err, os.ErrDeadlineExceeded) {
				if err := conn.(*net.TCPConn).SetLinger(0); err != nil {
					errChan <- fmt.Errorf("TCP connection reset: %w", err)
				}
			} else {
				errChan <- fmt.Errorf("abnormal closure: %w", err)

				now := time.Now().Add(time.Second * -1)
				if err := dest.SetReadDeadline(now); err != nil {
					errChan <- fmt.Errorf("read from the destination: %w", err)
				}
				if err := src.SetWriteDeadline(now); err != nil {
					errChan <- fmt.Errorf("write to the source: %w", err)
				}
			}
		}

		buf := make([]byte, 1024)
		for isEof := false; !isEof; {
			if n, err := src.Read(buf); err != nil {
				if err != io.EOF {
					reportError(src, err)
					break
				}
				isEof = true
			} else if _, err := dest.Write(buf[:n]); err != nil {
				reportError(dest, err)
				break
			}
		}
	}

	go transfer(clientConn, proxyConn)
	transfer(proxyConn, clientConn)
}
