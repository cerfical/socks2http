package http

import (
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox"
)

type requestHandler struct {
	clientConn net.Conn
	request    *http.Request
	proxy      *prox.Proxy
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
	if h.proxy.Addr().Scheme() == addr.HTTP {
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
