package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"socks2http/internal/args"
	"socks2http/internal/log"
	"socks2http/internal/proxy"
	"socks2http/internal/util"
	"sync"
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

	clientConn, err := getRawConnection(wr)
	if err != nil {
		return err
	}
	defer clientConn.Close()

	_, err = io.Copy(clientConn, conn)
	return err
}

func setupHTTPTunnel(wr http.ResponseWriter, proxyConn net.Conn) error {
	wr.WriteHeader(http.StatusOK)
	clientConn, err := getRawConnection(wr)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	transfer := func(dest io.WriteCloser, src io.ReadCloser) {
		if _, err := io.Copy(dest, src); err != nil {
			log.Error("HTTP tunnel closed: %v", err)
		}
		wg.Done()
	}

	go transfer(clientConn, proxyConn)
	transfer(proxyConn, clientConn)
	wg.Wait()

	util.Must(clientConn.Close())
	util.Must(proxyConn.Close())
	return nil
}

func getRawConnection(wr http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijacking not supported")
	}

	conn, _, err := hijacker.Hijack()
	return conn, err
}
