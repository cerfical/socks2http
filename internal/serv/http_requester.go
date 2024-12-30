package serv

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

type httpRequester struct{}

func (httpRequester) request(cliConn net.Conn) (request, error) {
	req, err := http.ReadRequest(bufio.NewReader(cliConn))
	if err != nil {
		return nil, err
	}

	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("parse the request URI: %v", err)
	}

	return &httpRequest{*destAddr, cliConn, req}, nil
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

type httpRequest struct {
	dest    addr.Addr
	cliConn net.Conn
	*http.Request
}

func (r *httpRequest) writeReply(ok bool) error {
	if r.Method == http.MethodConnect {
		resp := http.Response{ProtoMajor: 1, ProtoMinor: 1}
		if ok {
			resp.StatusCode = http.StatusOK
		} else {
			resp.StatusCode = http.StatusForbidden
		}
		return resp.Write(r.cliConn)
	}
	return nil
}

func (r *httpRequest) perform(proto string, servConn net.Conn, log *log.Logger) {
	log.WithAttrs(
		"method", r.Method,
		"uri", r.RequestURI,
		"proto", r.Proto,
	).Infof("incoming request")

	if r.Method == http.MethodConnect {
		for err := range tunnel(r.cliConn, servConn) {
			log.Errorf("%v", err)
		}
	} else {
		if err := r.forwardRequest(proto, servConn); err != nil {
			log.Errorf("%v", err)
		}
	}
}

func (r *httpRequest) forwardRequest(proto string, servConn net.Conn) error {
	// if the connection goes through an HTTP proxy
	if proto == addr.HTTP {
		// write the request as expected by the proxy
		if err := r.WriteProxy(servConn); err != nil {
			return err
		}
	} else {
		// otherwise just forward the request
		if err := r.Write(servConn); err != nil {
			return err
		}
	}

	_, err := io.Copy(r.cliConn, servConn)
	return err
}

func (r *httpRequest) destAddr() *addr.Addr {
	return &r.dest
}

func (r *httpRequest) Close() error {
	return r.Body.Close()
}
