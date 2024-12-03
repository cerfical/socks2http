package http

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"socks2http/internal/addr"
	"socks2http/internal/log"
	"socks2http/internal/prox"
)

func Run(servHost addr.Host, proxy prox.Proxy, logger log.Logger) error {
	server := httpServer{proxy, logger}
	if err := http.ListenAndServe(servHost.String(), &server); err != nil {
		return err
	}
	return nil
}

type httpServer struct {
	proxy  prox.Proxy
	logger log.Logger
}

func (s *httpServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := req.Method + " " + req.RequestURI + " " + req.Proto
	s.logger.Info(requestLine)

	clientConn, _, err := wr.(http.Hijacker).Hijack()
	if err != nil {
		s.logger.Error("opening a client connection: %v", err)
		return
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			s.logger.Error("closing client connection: %v", err)
		}
	}()

	s.handleRequest(clientConn, req)
}

func (s *httpServer) handleRequest(clientConn net.Conn, req *http.Request) {
	destHost, err := extractHost(req.URL)
	if err != nil {
		s.logger.Error("extracting destination host from %q: %v", req.URL, err)
		return
	}

	servConn, err := s.proxy.Open(destHost)
	if err != nil {
		s.logger.Error("opening proxy to %v: %v", destHost, err)
		return
	}
	defer func() {
		if err := servConn.Close(); err != nil {
			s.logger.Error("closing proxy connection: %v", err)
		}
	}()

	if req.Method == http.MethodConnect {
		for err := range tunnel(clientConn, servConn) {
			s.logger.Error("tunnel to %v: %v", destHost, err)
		}
	} else {
		if err := sendRequest(clientConn, servConn, req); err != nil {
			s.logger.Error("sending request to %v: %v", destHost, err)
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
