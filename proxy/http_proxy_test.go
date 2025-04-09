package proxy_test

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/suite"
)

const (
	megabyte = 1 << 20

	numInRowRequests = 10
	maxPayloadSize   = megabyte * 10
)

func TestHTTPProxy(t *testing.T) {
	suite.Run(t, new(HTTPProxyTest))
}

type HTTPProxyTest struct {
	ProxyTest
}

func (t *HTTPProxyTest) TestServe_HTTP() {
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

func (t *HTTPProxyTest) TestServe_HTTPS() {
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

func (t *HTTPProxyTest) TestConnect() {
	t.Run("make an HTTP CONNECT request to the proxy", func() {
		dstHost := addr.NewHost("localhost", 9999)
		proxyServerConn := t.proxyConnect(
			addr.New(addr.HTTP, "localhost", 8080),
			dstHost,
		)

		// Assert that the proxy server receives the correct request
		req, err := http.ReadRequest(bufio.NewReader(proxyServerConn))
		t.Require().NoError(err)

		t.Equal(dstHost.String(), req.Host)
		t.Equal(http.MethodConnect, req.Method)

		// Write a 200 OK response
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)

		t.Require().NoError(resp.Result().Write(proxyServerConn))
	})
}

func newTLSConn(c net.Conn) net.Conn {
	return tls.Client(c, &tls.Config{InsecureSkipVerify: true})
}

func makeString(size int) string {
	return strings.Repeat("abcd", size/4)
}
