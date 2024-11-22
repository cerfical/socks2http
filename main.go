package main

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"socks2http/internal/args"
	"socks2http/internal/proxy"
	"socks2http/internal/util"
	"sync"
	"time"
)

func main() {
	args := args.Parse()
	server := httpProxyServer{
		proxy:   proxy.NewProxy(args.Proxy.Host, args.Proxy.Proto, args.Timeout),
		timeout: args.Timeout,
	}

	switch args.Server.Proto {
	case "http":
		if err := http.ListenAndServe(args.Server.Host, &server); err != nil {
			util.FatalError("closing the server: %v", err)
		}
	default:
		util.FatalError("unsupported server protocol scheme: %v", args.Server.Proto)
	}
}

type httpProxyServer struct {
	proxy   proxy.Proxy
	timeout time.Duration
}

func (s *httpProxyServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	log.Println(requestLine)

	proxyConn, err := proxy.OpenURL(s.proxy, req.URL)
	if err != nil {
		log.Printf("failed to open up a proxy: %v\n", err)
		return
	}

	if req.Method != http.MethodConnect {
		if err := sendRequest(wr, req, proxyConn); err != nil {
			log.Printf("failed to use a proxy: %v\n", err)
		}
	} else {
		if err := setupHTTPTunnel(wr, proxyConn); err != nil {
			log.Printf("failed to setup an HTTP tunnel: %v\n", err)
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

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		transfer := func(dest io.WriteCloser, src io.ReadCloser) {
			if _, err := io.Copy(dest, src); err != nil {
				log.Printf("abnormal closure of an HTTP tunnel: %v\n", err)
			}
			wg.Done()
		}

		go transfer(clientConn, proxyConn)
		go transfer(proxyConn, clientConn)
		wg.Wait()

		util.Must(clientConn.Close())
		util.Must(proxyConn.Close())
	}()
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
