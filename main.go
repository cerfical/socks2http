package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"socks2http/socks"
	"strconv"
)

func main() {
	server := HttpProxyServer{}

	httpProxy := flag.String("http-proxy", "localhost:8080", "IP and port to run HTTP proxy server")
	flag.StringVar(&server.socksProxy, "socks-proxy", "localhost:1080", "IP and port of SOCKS proxy server to connect to")
	flag.Parse()

	if err := http.ListenAndServe(*httpProxy, &server); err != nil {
		log.Println(err)
	}
}

type HttpProxyServer struct {
	socksProxy string
}

func (s *HttpProxyServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	log.Println(requestLine)

	var err error
	if req.Method == http.MethodConnect {
		err = errors.New("unsupported method: " + req.Method)
	} else {
		err = s.setupHttpConnection(wr, req)
	}

	if err != nil {
		log.Println(err)
	}
}

func (s *HttpProxyServer) setupHttpConnection(wr http.ResponseWriter, req *http.Request) error {
	ip, err := net.ResolveIPAddr("ip", req.URL.Hostname())
	if err != nil {
		return err
	}

	proxyConn, err := net.Dial("tcp", s.socksProxy)
	if err != nil {
		return err
	}
	defer proxyConn.Close()

	if port, err := parsePort(req.URL.Port(), req.URL.Scheme); err != nil {
		return err
	} else if err := socks.Connect(proxyConn, ip.IP, port); err != nil {
		return fmt.Errorf("failed to connect to SOCKS4 proxy %v:%v: %v", ip.IP, port, err)
	} else if err := req.Write(proxyConn); err != nil {
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

func parsePort(port string, scheme string) (uint16, error) {
	if portNum, err := strconv.ParseUint(port, 10, 16); err == nil {
		return uint16(portNum), nil
	} else if scheme == "http" {
		return 80, nil
	} else if scheme == "https" {
		return 443, nil
	}
	return 0, errors.New("unsupported protocol scheme: " + scheme)
}

func getRawConnection(wr http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijacking not supported")
	}

	conn, _, err := hijacker.Hijack()
	return conn, err
}
