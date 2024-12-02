package http

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"socks2http/internal/proxy"
)

type HTTPServer struct {
	Host   addr.Host
	Proxy  proxy.Proxy
	Logger log.Logger
}

func (s *HTTPServer) Run() error {
	if err := http.ListenAndServe(s.Host.String(), s); err != nil {
		return err
	}
	return nil
}

func (s *HTTPServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.URL.String() + " " + req.Proto
	s.Logger.Info(requestLine)

	clientConn, _, err := wr.(http.Hijacker).Hijack()
	if err != nil {
		s.Logger.Error("opening a client connection: %v", err)
		return
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			s.Logger.Error("closing client connection: %v", err)
		}
	}()

	s.handleRequest(clientConn, req)
}

func (s *HTTPServer) handleRequest(clientConn net.Conn, req *http.Request) {
	destHost, err := extractHost(req.URL)
	if err != nil {
		s.Logger.Error("extracting destination host from %q: %v", req.URL, err)
		return
	}

	servConn, err := s.Proxy.Open(destHost)
	if err != nil {
		s.Logger.Error("opening proxy to %v: %v", destHost, err)
		return
	}
	defer func() {
		if err := servConn.Close(); err != nil {
			s.Logger.Error("closing proxy connection: %v", err)
		}
	}()

	if req.Method == http.MethodConnect {
		for err := range tunnel(clientConn, servConn) {
			s.Logger.Error("tunnel to %v: %v", destHost, err)
		}
	} else {
		if err := sendRequest(clientConn, servConn, req); err != nil {
			s.Logger.Error("sending request to %v: %v", destHost, err)
		}
	}
}

func extractHost(url *url.URL) (addr.Host, error) {
	port, err := makePort(url.Port(), url.Scheme)
	if err != nil {
		return addr.Host{}, err
	}
	return addr.Host{Hostname: url.Hostname(), Port: port}, nil
}

func makePort(portStr, scheme string) (uint16, error) {
	if portStr == "" {
		port, err := net.LookupPort("tcp", scheme)
		if err != nil {
			return 0, err
		}
		return uint16(port), nil
	}
	return addr.ParsePort(portStr)
}

func sendRequest(clientConn, servConn net.Conn, req *http.Request) error {
	if err := req.Write(servConn); err != nil {
		return err
	}
	_, err := io.Copy(clientConn, servConn)
	return err
}