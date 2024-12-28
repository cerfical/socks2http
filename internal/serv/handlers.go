package serv

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/cli"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/socks"
)

func handleHTTPRequest(cliConn net.Conn, prox *cli.ProxyClient, log *log.Logger) {
	handler := requestHandler{cliConn, prox, log}
	handler.run()
}

type requestHandler struct {
	cliConn net.Conn
	prox    *cli.ProxyClient
	log     *log.Logger
}

func (h *requestHandler) run() {
	req, err := http.ReadRequest(bufio.NewReader(h.cliConn))
	if err != nil {
		h.log.Errorf("request parsing: %v", err)
		return
	}

	defer func() {
		if err := req.Body.Close(); err != nil {
			h.log.Errorf("cleaning up request data: %v", err)
		}
	}()

	h.log.WithAttrs(
		"method", req.Method,
		"uri", req.RequestURI,
		"proto", req.Proto,
	).Infof("incoming request")

	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		h.log.Errorf("parsing request URI: %v", err)
		return
	}

	servConn, err := h.prox.Open(destAddr)
	if err != nil {
		h.log.Errorf("opening a server connection: %v", err)
		return
	}

	defer func() {
		if err := servConn.Close(); err != nil {
			h.log.Errorf("closing a server connection: %v", err)
		}
	}()

	if req.Method == http.MethodConnect {
		okResp := http.Response{StatusCode: http.StatusOK, ProtoMajor: 1, ProtoMinor: 1}
		if err := okResp.Write(h.cliConn); err != nil {
			h.log.Errorf("%v", err)
			return
		}

		for err := range tunnel(h.cliConn, servConn) {
			h.log.Errorf("%v", err)
		}
	} else {
		if err := h.forwardRequest(req, servConn); err != nil {
			h.log.Errorf("%v", err)
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

	p, err := addr.ParsePort(port)
	if err != nil {
		return nil, err
	}

	return &addr.Addr{
		Scheme:   url.Scheme,
		Hostname: url.Hostname(),
		Port:     p,
	}, nil
}

func (h *requestHandler) forwardRequest(req *http.Request, servConn net.Conn) error {
	// if the connection goes through an HTTP proxy
	if h.prox.Proto() == addr.HTTP {
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

	_, err := io.Copy(h.cliConn, servConn)
	return err
}

func handleSOCKS4Request(cliConn net.Conn, prox *cli.ProxyClient, log *log.Logger) {
	req, err := socks.ReadRequest(cliConn)
	if err != nil {
		log.Errorf("%v", err)
		return
	}

	addr := addr.Addr{Hostname: req.DestIP.String(), Port: req.DestPort}
	servConn, err := prox.Open(&addr)
	if err != nil {
		errRep := socks.Reply{Code: socks.RequestRejectedOrFailed}
		if err := errRep.Write(cliConn); err != nil {
			log.Errorf("%v", err)
		}
		log.Errorf("open a server connection: %v", err)
		return
	}

	defer func() {
		if err := servConn.Close(); err != nil {
			log.Errorf("close a server connection: %v", err)
		}
	}()

	okRep := socks.Reply{Code: socks.RequestGranted}
	if err := okRep.Write(cliConn); err != nil {
		log.Errorf("%v", err)
		return
	}

	for err := range tunnel(cliConn, servConn) {
		log.Errorf("%v", err)
	}
}
