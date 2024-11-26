package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/proxy"
	"sync"
	"time"
)

func main() {
	switch args.Server.Scheme {
	case "http":
		if err := http.ListenAndServe(args.Server.Host(), &httpProxyServer{}); err != nil {
			log.Fatal("unexpected server shutdown: %v", err)
		}
	default:
		log.Fatal("unsupported server protocol scheme %q", args.Server.Scheme)
	}
}

type httpProxyServer struct{}

func (s *httpProxyServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	log.Info(requestLine)

	proxyConn, err := proxy.Open(req.URL)
	if err != nil {
		log.Error("failed to proxy %v: %v", req.URL, err)
		return
	}

	if req.Method != http.MethodConnect {
		if err := sendRequest(wr, req, proxyConn); err != nil {
			log.Error("communication failed with %v: %v", req.URL, err)
		}
	} else {
		if err := setupHTTPTunnel(wr, proxyConn); err != nil {
			log.Error("failed to setup an HTTP tunnel to %v: %v", req.URL, err)
		}
	}
}

func sendRequest(wr http.ResponseWriter, req *http.Request, conn net.Conn) error {
	defer conn.Close()
	if err := req.Write(conn); err != nil {
		return err
	}

	clientConn, err := rawConn(wr)
	if err != nil {
		return err
	}
	defer clientConn.Close()

	_, err = io.Copy(clientConn, conn)
	return err
}

func setupHTTPTunnel(wr http.ResponseWriter, proxyConn net.Conn) error {
	defer proxyConn.Close()

	wr.WriteHeader(http.StatusOK)
	clientConn, err := rawConn(wr)
	if err != nil {
		return err
	}
	defer clientConn.Close()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	transfer := func(dest, src net.Conn) {
		defer wg.Done()
		wg.Add(1)

		reportError := func(conn net.Conn, err error) {
			// use deadlines to preemptively terminate Read()/Write() calls and avoid goroutines being blocked indefinitely
			if errors.Is(err, os.ErrDeadlineExceeded) {
				if err := conn.(*net.TCPConn).SetLinger(0); err != nil {
					log.Error("failed to reset TCP connection after error: %v", err)
				}
			} else {
				now := time.Now().Add(time.Second * -1)
				if err := dest.SetReadDeadline(now); err != nil {
					log.Error("failed to close HTTP tunnel: %v", err)
				}
				if err := src.SetWriteDeadline(now); err != nil {
					log.Error("failed to close HTTP tunnel: %v", err)
				}
				log.Error("HTTP tunnel closed abnormally: %v", err)
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
	return nil
}

func rawConn(wr http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijacking not supported")
	}

	conn, _, err := hijacker.Hijack()
	return conn, err
}
