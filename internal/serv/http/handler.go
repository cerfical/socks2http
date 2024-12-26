package http

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/cli"
	"github.com/cerfical/socks2http/internal/log"
)

func NewHandler(prox *cli.ProxyClient, log *log.Logger) *Handler {
	return &Handler{prox, log}
}

type Handler struct {
	prox *cli.ProxyClient
	log  *log.Logger
}

func (h *Handler) Handle(cliConn net.Conn) {
	req, err := http.ReadRequest(bufio.NewReader(cliConn))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			h.log.Errorf("parsing HTTP request: %v", err)
		}
		return
	}

	log := h.log.WithAttr("uri", req.RequestURI)
	log.WithAttrs(
		"method", req.Method,
		"proto", req.Proto,
	).Infof("new request")

	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		log.Errorf("parsing destination server address: %v", err)
		return
	}

	servConn, err := h.prox.Open(destAddr)
	if err != nil {
		log.Errorf("opening a server connection: %v", err)
		return
	}

	defer func() {
		if err := servConn.Close(); err != nil {
			log.Errorf("closing a server connection: %v", err)
		}
	}()

	if req.Method == http.MethodConnect {
		for err := range tunnel(cliConn, servConn) {
			log.Errorf("HTTP tunneling: %v", err)
		}
	} else {
		if err := h.forwardRequest(req, cliConn, servConn); err != nil {
			log.Errorf("%v", err)
		}
	}
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

	return &addr.Addr{
		Scheme:   url.Scheme,
		Hostname: url.Hostname(),
		Port:     port,
	}, nil
}

func (h *Handler) forwardRequest(req *http.Request, cliConn, servConn net.Conn) error {
	// if the connection goes through an HTTP proxy
	if h.prox.Addr().Scheme == addr.HTTP {
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

	_, err := io.Copy(cliConn, servConn)
	return err
}
