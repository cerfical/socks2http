package http

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox/cli"
)

type requestHandler struct {
	clientConn net.Conn
	request    *http.Request
	proxy      *cli.Proxy
	logger     *log.Logger
}

func (h *requestHandler) run() {
	destAddr, err := addrFromURL(h.request.URL)
	if err != nil {
		h.logger.Errorf("parsing destination server address: %v", err)
		return
	}

	servConn, err := h.proxy.Open(destAddr)
	if err != nil {
		h.logger.Errorf("opening server connection: %v", err)
		return
	}

	defer func() {
		if err := servConn.Close(); err != nil {
			h.logger.Errorf("closing server connection: %v", err)
		}
	}()

	if h.request.Method == http.MethodConnect {
		for err := range tunnel(h.clientConn, servConn) {
			h.logger.Errorf("HTTP tunneling: %v", err)
		}
	} else {
		if err := h.sendRequest(servConn); err != nil {
			h.logger.Errorf("%v", err)
		}
	}
}

func (h *requestHandler) sendRequest(servConn net.Conn) error {
	// if the connection goes through an HTTP proxy
	if h.proxy.Addr().Scheme == addr.HTTP {
		// write the request as expected by the proxy
		if err := h.request.WriteProxy(servConn); err != nil {
			return err
		}
	} else {
		// otherwise just forward the request
		if err := h.request.Write(servConn); err != nil {
			return err
		}
	}

	_, err := io.Copy(h.clientConn, servConn)
	return err
}

func addrFromURL(url *url.URL) (*addr.Addr, error) {
	port := url.Port()
	if port == "" {
		p, err := net.LookupPort("tcp", url.Scheme)
		if err != nil {
			return nil, err
		}
		port = strconv.Itoa(p)
	}
	return &addr.Addr{Scheme: url.Scheme, Hostname: url.Hostname(), Port: port}, nil
}
