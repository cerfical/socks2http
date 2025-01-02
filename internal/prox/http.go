package prox

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

type httpHandler struct{}

func (httpHandler) parseRequest(r *bufio.Reader) (request, error) {
	req, err := http.ReadRequest(r)
	if err != nil {
		return nil, err
	}

	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("parse request URI: %w", err)
	}

	return &httpRequest{req, *destAddr}, nil
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
	*http.Request
	dest addr.Addr
}

func (r *httpRequest) isConnect() bool {
	return r.Method == http.MethodConnect
}

func (r *httpRequest) destAddr() *addr.Addr {
	return &r.dest
}

func (r *httpRequest) logAttrs(log *log.Logger) *log.Logger {
	return log.WithAttrs(
		"method", r.Method,
		"uri", r.RequestURI,
		"proto", r.Proto,
	)
}

func (r *httpRequest) writeGrant(w io.Writer) error {
	return r.writeReply(w, true)
}

func (r *httpRequest) writeReject(w io.Writer) error {
	return r.writeReply(w, false)
}

func (r *httpRequest) writeReply(w io.Writer, ok bool) error {
	resp := http.Response{ProtoMajor: 1, ProtoMinor: 1}
	if ok {
		resp.StatusCode = http.StatusOK
	} else {
		resp.StatusCode = http.StatusForbidden
	}
	return resp.Write(w)
}

func (r *httpRequest) write(w io.Writer) error {
	return r.Write(w)
}

func (r *httpRequest) writeProxy(w io.Writer) error {
	return r.WriteProxy(w)
}

func (r *httpRequest) Close() (err error) {
	defer func() {
		if errClose := r.Body.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()

	// discard the request body if there is any left
	if _, err := io.ReadAll(r.Body); err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	return nil
}
