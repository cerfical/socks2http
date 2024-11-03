package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"socks2http/socks"
	"strconv"
	"sync"
	"time"
)

func main() {
	server := HttpProxyServer{}

	httpProxy := flag.String("http-proxy", "localhost:8080", "IP and port to run HTTP proxy server")
	flag.StringVar(&server.socksProxy, "socks-proxy", "localhost:1080", "IP and port of SOCKS proxy server to connect to")
	flag.DurationVar(&server.timeout, "timeout", 0, "time to wait for connection, no timeout by default (0)")
	flag.Parse()

	if err := http.ListenAndServe(*httpProxy, &server); err != nil {
		log.Println(err)
	}
}

type HttpProxyServer struct {
	socksProxy string
	timeout    time.Duration
}

func (s *HttpProxyServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	log.Println(requestLine)

	var err error
	if req.Method == http.MethodConnect {
		err = s.setupTunnel(wr, req)
	} else {
		err = s.setupHttpConnection(wr, req)
	}

	if err != nil {
		log.Println(err)
	}
}

func (s *HttpProxyServer) setupHttpConnection(wr http.ResponseWriter, req *http.Request) error {
	proxyConn, err := s.dialProxy(req)
	if err != nil {
		return err
	}
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

func (s *HttpProxyServer) setupTunnel(wr http.ResponseWriter, req *http.Request) error {
	proxyConn, err := s.dialProxy(req)
	if err != nil {
		return err
	}

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
				log.Println(err)
			}
			wg.Done()
		}

		go transfer(clientConn, proxyConn)
		go transfer(proxyConn, clientConn)
		wg.Wait()

	}()
	return nil
}

func (s *HttpProxyServer) dialProxy(req *http.Request) (net.Conn, error) {
	destAddr, err := url2Addr(req.URL)
	if err != nil {
		return nil, err
	}

	proxyConn, err := socks.DialTimeout(s.socksProxy, destAddr, s.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to proxy %v: %w", destAddr, err)
	}
	return proxyConn, err
}

func url2Addr(url *url.URL) (string, error) {
	host := url.Hostname()
	port := url.Port()

	if port == "" {
		portNum, err := net.LookupPort("tcp", url.Scheme)
		if err != nil {
			return "", err
		}
		port = strconv.Itoa(portNum)
	}
	return host + ":" + port, nil
}

func getRawConnection(wr http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijacking not supported")
	}

	conn, _, err := hijacker.Hijack()
	return conn, err
}
