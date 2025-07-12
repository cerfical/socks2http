package server

import (
	stdlog "log"
	"sync"

	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
)

type HTTPServer struct {
	Tunneler proxy.Tunneler
	Dialer   proxy.Dialer

	Log proxy.Logger

	activeTunnels sync.WaitGroup
}

func (s *HTTPServer) ServeHTTP(ctx context.Context, l net.Listener) error {
	server := http.Server{
		Handler:  http.HandlerFunc(s.handle),
		ErrorLog: stdlog.New(httpErrorLog{s}, "", 0),
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(l)
	}()

	select {
	case <-ctx.Done():
		err := server.Shutdown(context.Background())
		s.activeTunnels.Wait()
		return err
	case err := <-errChan:
		return err
	}
}

func (s *HTTPServer) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.connect(w, r)
	} else {
		s.forwardRequest(w, r)
	}
}

func (s *HTTPServer) connect(w http.ResponseWriter, r *http.Request) {
	dstAddr, err := hostFromHTTPConnect(r)
	if err != nil {
		s.httpStatus(w, r, http.StatusBadRequest, fmt.Errorf("parse destination address: %w", err))
		return
	}

	dstConn, err := s.Dialer.Dial(r.Context(), dstAddr)
	if err != nil {
		s.httpStatus(w, r, http.StatusBadGateway, fmt.Errorf("connect to destination: %w", err))
		return
	}
	defer dstConn.Close()

	// Take over the client connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		s.httpStatus(w, r, http.StatusInternalServerError, fmt.Errorf("hijacker not supported"))
		return
	}

	s.activeTunnels.Add(1)
	defer s.activeTunnels.Done()

	clientConn, _, err := hj.Hijack()
	if err != nil {
		s.httpStatus(w, r, http.StatusInternalServerError, fmt.Errorf("hijack connection: %w", err))
		return
	}
	defer clientConn.Close()

	// Establish the HTTP tunnel
	if !s.httpStatus(clientConn, r, http.StatusOK, nil) {
		return
	}
	if err := s.Tunneler.Tunnel(r.Context(), clientConn, dstConn); err != nil {
		s.serverError(fmt.Errorf("proxy tunnel: %w", err))
		return
	}
}

func (s *HTTPServer) forwardRequest(w http.ResponseWriter, r *http.Request) {
	dstAddr, err := hostFromHTTPRequest(r)
	if err != nil {
		s.httpStatus(w, r, http.StatusBadRequest, fmt.Errorf("parse destination address: %w", err))
		return
	}

	dstConn, err := s.Dialer.Dial(r.Context(), dstAddr)
	if err != nil {
		s.httpStatus(w, r, http.StatusBadGateway, fmt.Errorf("connect to destination: %w", err))
		return
	}
	defer dstConn.Close()

	if err := r.Write(dstConn); err != nil {
		s.httpStatus(w, r, http.StatusBadGateway, fmt.Errorf("write request: %w", err))
		return
	}

	// Forward the response from the destination to the client
	resp, err := http.ReadResponse(bufio.NewReader(dstConn), r)
	if err != nil {
		s.httpStatus(w, r, http.StatusBadGateway, fmt.Errorf("read response: %w", err))
		return
	}
	defer resp.Body.Close()

	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(header, value)
		}
	}
	s.httpStatus(w, r, resp.StatusCode, nil)

	if _, err := io.Copy(w, resp.Body); err != nil {
		s.serverError(fmt.Errorf("copy response body: %w", err))
		return
	}
}

func (s *HTTPServer) httpStatus(w io.Writer, r *http.Request, status int, err error) bool {
	msg := fmt.Sprintf("%v %v", r.Method, r.RequestURI)
	fields := []any{
		"proto", r.Proto,
		"status", fmt.Sprintf("%v %v", status, http.StatusText(status)),
		"client", r.RemoteAddr,
	}

	// Log the request and response status
	if err != nil {
		s.Log.Error(msg, append(fields,
			"error", err,
		)...)
	} else {
		s.Log.Info(msg, fields...)
	}

	// Write the status code to the client
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.WriteHeader(status)
	} else {
		response := http.Response{
			StatusCode: status,
			ProtoMajor: r.ProtoMajor,
			ProtoMinor: r.ProtoMinor,
		}
		if err := response.Write(w); err != nil {
			s.serverError(fmt.Errorf("write response: %w", err))
			return false
		}
	}

	return true
}

func (s *HTTPServer) serverError(err error) {
	s.Log.Error("HTTP failure", "error", err)
}

type httpErrorLog struct {
	server *HTTPServer
}

func (w httpErrorLog) Write(p []byte) (int, error) {
	// Trim carriage return produced by stdlog
	n := len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[0 : n-1]
		n--
	}

	w.server.serverError(fmt.Errorf("%v", string(p)))
	return n, nil
}

func hostFromHTTPConnect(r *http.Request) (*addr.Addr, error) {
	h, err := addr.ParseAddr(r.URL.Host)
	if err != nil {
		return nil, fmt.Errorf("parse host: %w", err)
	}
	return h, nil
}

func hostFromHTTPRequest(r *http.Request) (*addr.Addr, error) {
	// For proxied requests, the request URI contains the full destination URL
	port := r.URL.Port()
	if port == "" {
		// If the URL contains no port, we can try to guess it by looking at the scheme
		portNum, err := net.LookupPort("tcp", r.URL.Scheme)
		if err != nil {
			return nil, fmt.Errorf("lookup port: %w", err)
		}
		return addr.NewAddr(r.URL.Hostname(), uint16(portNum)), nil
	}

	// If the port is specified, we can use it directly
	portNum, err := addr.ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}

	return addr.NewAddr(r.URL.Hostname(), portNum), nil
}
