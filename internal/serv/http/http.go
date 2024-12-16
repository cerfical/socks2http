package http

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox"
)

func Run(servAddr *addr.Addr, proxy *prox.Proxy, logger *log.Logger) error {
	server := httpServer{proxy, logger}
	if err := http.ListenAndServe(servAddr.Host(), &server); err != nil {
		return err
	}
	return nil
}

type httpServer struct {
	proxy  *prox.Proxy
	logger *log.Logger
}

func (s *httpServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	requestLine := fmt.Sprintf("%v %v %v", req.Method, req.RequestURI, req.Proto)
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
	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		s.logger.Error("parsing destination %v: %v", req.URL, err)
		return
	}

	servConn, err := s.proxy.Open(destAddr)
	if err != nil {
		s.logger.Error("opening proxy to %v: %v", destAddr, err)
		return
	}
	defer func() {
		if err := servConn.Close(); err != nil {
			s.logger.Error("closing proxy %v: %v", s.proxy.Addr(), err)
		}
	}()

	if req.Method == http.MethodConnect {
		for err := range tunnel(clientConn, servConn) {
			s.logger.Error("tunneling %v: %v", destAddr, err)
		}
	} else {
		if err := s.sendRequest(clientConn, servConn, req); err != nil {
			s.logger.Error("requesting %v: %v", destAddr, err)
		}
	}
}

func (s *httpServer) sendRequest(clientConn, servConn net.Conn, req *http.Request) error {
	// if the connection goes through an HTTP proxy
	if s.proxy.Addr().Scheme() == addr.HTTP {
		// write the request as expected by the proxy
		if err := req.WriteProxy(servConn); err != nil {
			return err
		}
	} else {
		// otherwise just forward the request
		if err := req.Write(servConn); err != nil {
			return err
		}
	}

	_, err := io.Copy(clientConn, servConn)
	return err
}

func addrFromURL(url *url.URL) (*addr.Addr, error) {
	var port uint16
	if p := url.Port(); p != "" {
		p, err := addr.ParsePort(p)
		if err != nil {
			return nil, err
		}
		port = p
	} else {
		p, err := net.LookupPort("tcp", url.Scheme)
		if err != nil {
			return nil, err
		}
		port = uint16(p)
	}
	return addr.New(url.Scheme, url.Hostname(), port), nil
}
