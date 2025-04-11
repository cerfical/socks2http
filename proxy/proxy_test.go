package proxy_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/suite"
)

const (
	megabyte = 1 << 20

	numInRowRequests = 10
	maxPayloadSize   = megabyte * 10
)

func TestProxy(t *testing.T) {
	suite.Run(t, new(ProxyTest))
}

type ProxyTest struct {
	suite.Suite
}

func (t *ProxyTest) TestServe_SOCKS() {
	serverHost := t.startHTTPEchoServer()
	tests := map[string]struct {
		dstHost   *addr.Host
		wantReply *socks.Reply
	}{
		"supports tunneling of HTTP requests": {
			dstHost:   serverHost,
			wantReply: socks.NewReply(socks.Granted),
		},

		"responds with a Request Rejected if the server is unreachable": {
			dstHost:   addr.NewHost("0.0.0.0", 0),
			wantReply: socks.NewReply(socks.Rejected),
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			// TODO: Ensure the proxy is closed
			proxyConn, _ := t.proxyServe(addr.New(addr.SOCKS4, "localhost", 1080))

			// Write a SOCKS CONNECT request
			req := socks.NewRequest(socks.V4, socks.Connect, test.dstHost)
			t.Require().NoError(req.Write(proxyConn))

			reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
			t.Require().NoError(err)

			t.Equal(test.wantReply, reply)
			if reply.Status == socks.Granted {
				// Make a bunch of HTTP requests
				for range numInRowRequests {
					msg := makeString(maxPayloadSize)

					req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(msg))
					t.Require().NoError(req.Write(proxyConn))
					resp := t.readHTTPResponse(proxyConn)

					t.Equal(http.StatusOK, resp.StatusCode)
					t.Equal(msg, t.readString(resp.Body))
				}
			}
		})
	}
}

func (t *ProxyTest) TestServe_HTTP() {
	httpHost := t.startHTTPEchoServer()
	tests := map[string]struct {
		dstHost    *addr.Host
		body       string
		wantStatus int
	}{
		"supports forwarding of an HTTP request": {
			dstHost:    httpHost,
			body:       "Everything is OK",
			wantStatus: http.StatusOK,
		},

		"supports forwarding of an HTTP request with large payload": {
			dstHost:    httpHost,
			body:       makeString(maxPayloadSize),
			wantStatus: http.StatusOK,
		},

		"responds with a 502 Bad Gateway if the server is unreachable": {
			dstHost:    addr.NewHost("0.0.0.0", 0),
			wantStatus: http.StatusBadGateway,
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			proxyConn, proxyErr := t.proxyServe(addr.New(addr.HTTP, "localhost", 8080))

			// Send the request to the proxy
			requestURI := fmt.Sprintf("http://%v", test.dstHost)
			req := httptest.NewRequest(http.MethodPost, requestURI, strings.NewReader(test.body))
			req.Close = true

			t.Require().NoError(req.WriteProxy(proxyConn))
			resp := t.readHTTPResponse(proxyConn)

			t.Equal(test.wantStatus, resp.StatusCode)
			t.Equal(test.body, t.readString(resp.Body))

			if err := <-proxyErr; test.wantStatus == http.StatusOK {
				t.NoError(err)
			} else {
				t.Error(err)
			}
		})
	}
}

func (t *ProxyTest) TestServe_HTTPS() {
	httpsHost := t.startHTTPSEchoServer()
	tests := map[string]struct {
		dstHost    *addr.Host
		makeBody   func(int) string
		wantStatus int
	}{
		"supports tunneling of HTTPS requests": {
			dstHost: httpsHost,
			makeBody: func(reqNum int) string {
				return fmt.Sprintf("Message #%v", reqNum+1)
			},
			wantStatus: http.StatusOK,
		},

		"supports tunneling of HTTPS requests with large payload": {
			dstHost: httpsHost,
			makeBody: func(int) string {
				return makeString(maxPayloadSize)
			},
			wantStatus: http.StatusOK,
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			// TODO: Ensure the proxy is properly closed
			proxyConn, _ := t.proxyServe(addr.New(addr.HTTP, "localhost", 8080))

			// Make an HTTP CONNECT request to the proxy
			req := httptest.NewRequest(http.MethodConnect, test.dstHost.String(), nil)
			t.Require().NoError(req.WriteProxy(proxyConn))

			resp := t.readHTTPResponse(proxyConn)
			t.Equal(http.StatusOK, resp.StatusCode)
			t.Equal("", t.readString(resp.Body))

			// Perform multiple HTTPS requests in a row through the established proxy tunnel
			proxyConn = newTLSConn(proxyConn)
			for i := range numInRowRequests {
				body := test.makeBody(i)

				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
				t.Require().NoError(req.Write(proxyConn))
				resp := t.readHTTPResponse(proxyConn)

				t.Equal(test.wantStatus, resp.StatusCode)
				t.Equal(body, t.readString(resp.Body))
			}
		})
	}
}

func (t *ProxyTest) proxyServe(proxyAddr *addr.Addr) (proxyConn net.Conn, proxyErr <-chan error) {
	proxy, err := proxy.New(&proxy.Options{
		Proto:  proxyAddr.Scheme,
		Dialer: proxy.Direct,
		Log:    log.Discard,
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

func newTLSConn(c net.Conn) net.Conn {
	return tls.Client(c, &tls.Config{InsecureSkipVerify: true})
}

func makeString(size int) string {
	return strings.Repeat("abcd", size/4)
}
