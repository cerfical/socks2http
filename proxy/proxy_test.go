package proxy_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/stretchr/testify/suite"
)

type ProxyTest struct {
	suite.Suite
}

func (t *ProxyTest) proxyServe(proxyAddr *addr.Addr) (proxyConn net.Conn, proxyErr <-chan error) {
	proxy, err := proxy.New(&proxy.Options{
		Addr:   *proxyAddr,
		Log:    log.Discard,
		Dialer: proxy.Direct,
	})
	t.Require().NoError(err)

	proxyClientConn, proxyServerConn := net.Pipe()
	t.T().Cleanup(func() {
		proxyClientConn.Close()
		proxyServerConn.Close()
	})

	errChan := make(chan error)
	go func() {
		errChan <- proxy.Serve(context.Background(), proxyServerConn)
	}()
	return proxyClientConn, errChan
}

func (t *ProxyTest) proxyConnect(proxyAddr *addr.Addr, dstHost *addr.Host) (proxyServerConn net.Conn) {
	proxyClientConn, proxyServerConn := net.Pipe()
	t.T().Cleanup(func() {
		proxyClientConn.Close()
		proxyServerConn.Close()
	})

	proxy, err := proxy.New(&proxy.Options{
		Addr: *proxyAddr,
		Log:  log.Discard,
		Dialer: proxy.DialerFunc(
			func(ctx context.Context, h *addr.Host) (net.Conn, error) {
				return proxyClientConn, nil
			},
		),
	})
	t.Require().NoError(err)

	errChan := make(chan error)
	go func() {
		_, err := proxy.Connect(context.Background(), dstHost)
		errChan <- err
	}()

	// Wait for the client to finish
	t.T().Cleanup(func() {
		t.Require().NoError(<-errChan)
	})

	return proxyServerConn
}

func (t *ProxyTest) readHTTPResponse(proxyConn net.Conn) *http.Response {
	t.T().Helper()

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
	t.Require().NoError(err)
	t.T().Cleanup(func() { resp.Body.Close() })

	return resp
}

func (t *ProxyTest) startHTTPEchoServer() *addr.Host {
	t.T().Helper()

	server := httptest.NewServer(http.HandlerFunc(echoHTTP))
	t.T().Cleanup(server.Close)

	return t.hostFromURL(server.URL)
}

func (t *ProxyTest) startHTTPSEchoServer() *addr.Host {
	t.T().Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(echoHTTP))
	t.T().Cleanup(server.Close)

	return t.hostFromURL(server.URL)
}

func (t *ProxyTest) hostFromURL(s string) *addr.Host {
	t.T().Helper()

	url, err := url.Parse(s)
	t.Require().NoError(err)

	h, err := addr.ParseHost(url.Host)
	t.Require().NoError(err)

	return h
}

func (t *ProxyTest) readString(r io.Reader) string {
	t.T().Helper()

	bytes, err := io.ReadAll(r)
	t.Require().NoError(err)

	return string(bytes)
}

func echoHTTP(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	w.Write(bytes)
}
