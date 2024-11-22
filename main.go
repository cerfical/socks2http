package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"socks2http/socks"
	"socks2http/util"
	"strconv"
	"sync"
	"time"
)

func main() {
	server := httpProxyServer{}

	httpProxy := flag.String("http-proxy", "localhost:8080", "IP and port to run HTTP proxy server")
	flag.StringVar(&server.socksProxy, "socks-proxy", "localhost:1080", "IP and port of SOCKS proxy server to connect to")
	flag.DurationVar(&server.timeout, "timeout", 0, "time to wait for connection, no timeout by default (0)")
	flag.Parse()

	if err := http.ListenAndServe(*httpProxy, &server); err != nil {
		log.Println(err)
	}
}

type httpProxyServer struct {
	socksProxy string
	timeout    time.Duration
}

func (s *httpProxyServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	log.Println(requestLine)

	destAddr, err := url2Addr(req.URL)
	if err != nil {
		log.Printf("failed to parse an URL %v: %v\n", req.URL, err)
		return
	}

	proxyConn, err := socks.ConnectTimeout(s.socksProxy, destAddr, s.timeout)
	if err != nil {
		log.Printf("failed to setup a proxy: %v\n", err)
		return
	}

	if req.Method != http.MethodConnect {
		if err := proxyRequest(wr, req, proxyConn); err != nil {
			log.Printf("failed to send a request via proxy: %v\n", err)
		}
	} else {
		if err := setupHTTPTunnel(wr, proxyConn, destAddr); err != nil {
			log.Printf("failed to setup an HTTP tunnel: %v\n", err)
		}
	}
}

func url2Addr(url *url.URL) (string, error) {
	port := url.Port()
	if port == "" {
		portNum, err := net.LookupPort("tcp", url.Scheme)
		if err != nil {
			return "", err
		}
		port = strconv.Itoa(portNum)
	}
	return url.Hostname() + ":" + port, nil
}

func proxyRequest(wr http.ResponseWriter, req *http.Request, proxyConn net.Conn) error {
	defer proxyConn.Close()
	if err := req.Write(proxyConn); err != nil {
		return err
	}

	clientConn, err := getRawConnection(wr)
	if err != nil {
		return err
	}
	defer clientConn.Close()

	_, err = io.Copy(clientConn, proxyConn)
	return err
}

func setupHTTPTunnel(wr http.ResponseWriter, proxyConn net.Conn, destAddr string) error {
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
				log.Printf("closing an HTTP tunnel for %v: %v\n", destAddr, err)
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
