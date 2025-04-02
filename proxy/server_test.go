package proxy_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/suite"
)

const (
	megabyte = 1 << 20

	numInRowRequests = 10
	maxPayloadSize   = megabyte * 10

	unreachableHost = "0.0.0.0:1234"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTest))
}

type ServerTest struct {
	ProxyTest
}

func (t *ServerTest) TestNew() {
	tests := map[string]struct {
		options proxy.ServerOption
		want    func(*proxy.Server)
		err     func(error)
	}{
		"uses http://localhost:8080 as the default listen address": {
			want: func(s *proxy.Server) {
				t.Equal(addr.New(addr.HTTP, "localhost", 8080), s.ListenAddr())
			},
		},

		"uses a non-default listen address if one is provided": {
			options: proxy.WithListenAddr(addr.New(addr.HTTP, "example.com", 8181)),
			want: func(s *proxy.Server) {
				t.Equal(addr.New(addr.HTTP, "example.com", 8181), s.ListenAddr())
			},
		},

		"rejects unsupported protocol schemes": {
			options: proxy.WithListenAddr(addr.New("SOCKS", "", 0)),
			err: func(err error) {
				t.ErrorContains(err, "SOCKS")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			ops := []proxy.ServerOption{}
			if test.options != nil {
				ops = append(ops, test.options)
			}

			serv, err := proxy.NewServer(ops...)
			if test.err != nil {
				test.err(err)
			} else {
				t.Require().NoError(err)
				test.want(serv)
			}
		})
	}
}

func (t *ServerTest) TestStart() {
	tests := map[string]struct {
		options proxy.ServerOption
		want    func(*proxy.Server)
	}{
		"starts to listen on the specified address": {
			options: proxy.WithListenAddr(addr.New(addr.HTTP, "localhost", 0)),
			want: func(s *proxy.Server) {
				t.assertHostIsReachable(s.ListenAddr().Host())
			},
		},

		"allocates a listen port if one was not provided": {
			options: proxy.WithListenAddr(addr.New(addr.HTTP, "localhost", 0)),
			want: func(s *proxy.Server) {
				t.NotZero(s.ListenAddr().Port)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			ops := []proxy.ServerOption{}
			if test.options != nil {
				ops = append(ops, test.options)
			}

			server, err := proxy.NewServer(ops...)
			t.Require().NoError(err)

			t.Require().NoError(server.Start(context.Background()))
			t.T().Cleanup(func() { server.Stop() })

			test.want(server)
		})
	}
}

func (t *ServerTest) TestServe_HTTP() {
	tests := map[string]struct {
		want func(net.Conn)
	}{
		"supports forwarding of a single HTTP request": {
			want: func(proxyConn net.Conn) {
				serverHost := t.startHTTPEchoServer()

				t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, "everything is OK", serverHost, proxyConn)
			},
		},

		"supports forwarding of a single HTTP request with a large payload": {
			want: func(proxyConn net.Conn) {
				serverHost := t.startHTTPEchoServer()

				largeMsg := makeString(maxPayloadSize)
				t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, largeMsg, serverHost, proxyConn)
			},
		},

		"supports tunneling of HTTPS requests": {
			want: func(proxyConn net.Conn) {
				serverHost := t.startHTTPSEchoServer()

				t.assertHTTPEchoesBack(http.StatusOK, http.MethodConnect, "", serverHost, proxyConn)

				proxyConn = tls.Client(proxyConn, &tls.Config{InsecureSkipVerify: true})
				for i := range numInRowRequests {
					msg := fmt.Sprintf("message #%v", i+1)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, msg, serverHost, proxyConn)
				}
			},
		},

		"supports tunneling of HTTPS requests with large payloads": {
			want: func(proxyConn net.Conn) {
				serverHost := t.startHTTPSEchoServer()

				t.assertHTTPEchoesBack(http.StatusOK, http.MethodConnect, "", serverHost, proxyConn)

				proxyConn = tls.Client(proxyConn, &tls.Config{InsecureSkipVerify: true})
				for range numInRowRequests {
					largeMsg := makeString(maxPayloadSize)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, largeMsg, serverHost, proxyConn)
				}
			},
		},

		"responds with a 502 Bad Gateway if the server is unreachable": {
			want: func(proxyConn net.Conn) {
				t.assertHTTPEchoesBack(http.StatusBadGateway, http.MethodConnect, "", unreachableHost, proxyConn)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			server := t.openProxyConn(addr.HTTP)
			test.want(server)
		})
	}
}

func (t *ServerTest) TestServe_SOCKS4() {
	tests := map[string]struct {
		want func(net.Conn)
	}{
		"supports tunneling of HTTP requests": {
			want: func(proxyConn net.Conn) {
				serverHost := t.startHTTPEchoServer()

				req := t.newSOCKS4Request(socks.RequestConnect, serverHost)
				reply := t.roundTripSOCKS4(req, proxyConn)
				t.Require().NoError(reply)

				for range numInRowRequests {
					msg := makeString(maxPayloadSize)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, msg, serverHost, proxyConn)
				}
			},
		},

		"responds with a Request Rejected if the server is unreachable": {
			want: func(proxyConn net.Conn) {
				req := t.newSOCKS4Request(socks.RequestConnect, unreachableHost)
				reply := t.roundTripSOCKS4(req, proxyConn)
				t.Require().ErrorContains(reply, "rejected")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			server := t.openProxyConn(addr.SOCKS4)
			test.want(server)
		})
	}
}

func (t *ServerTest) assertHostIsReachable(host string) {
	t.T().Helper()

	conn, err := net.Dial("tcp", host)
	t.NoError(err)
	t.T().Cleanup(func() { conn.Close() })
}

func (t *ServerTest) assertHTTPEchoesBack(status int, method, msg, serverHost string, serverConn net.Conn) {
	t.T().Helper()

	req := t.newHTTPRequest(method, serverHost, msg)
	resp := t.roundTripHTTP(req, serverConn)

	t.Equal(status, resp.StatusCode)
	t.Equal(msg, t.readString(resp.Body))
}

func (t *ServerTest) newHTTPRequest(method, host, body string) *http.Request {
	t.T().Helper()

	r, err := http.NewRequest(method, "", strings.NewReader(body))
	t.Require().NoError(err)
	r.Host = host

	return r
}

func (t *ServerTest) roundTripHTTP(r *http.Request, serverConn net.Conn) *http.Response {
	t.T().Helper()

	t.Require().NoError(r.Write(serverConn))

	resp, err := http.ReadResponse(bufio.NewReader(serverConn), r)
	t.Require().NoError(err)

	return resp
}

func (t *ServerTest) newSOCKS4Request(cmd byte, host string) *socks.Request {
	t.T().Helper()

	hostname, port, err := net.SplitHostPort(host)
	t.Require().NoError(err)

	portNum, err := addr.ParsePort(port)
	t.Require().NoError(err)

	ipAddr, err := addr.LookupIPv4(hostname)
	t.Require().NoError(err)

	return &socks.Request{
		Header: socks.Header{
			Version:  socks.V4,
			Command:  cmd,
			DestIP:   ipAddr,
			DestPort: portNum,
		},
		User: "",
	}
}

func (t *ServerTest) roundTripSOCKS4(r *socks.Request, serverConn net.Conn) error {
	t.T().Helper()

	t.Require().NoError(r.Write(serverConn))
	return socks.ReadReply(serverConn)
}

func (t *ServerTest) readString(r io.Reader) string {
	t.T().Helper()

	bytes, err := io.ReadAll(r)
	t.Require().NoError(err)
	return string(bytes)
}

func (t *ServerTest) openProxyConn(proto string) (proxyConn net.Conn) {
	t.T().Helper()

	server := t.startProxyServer(proto)
	proxyConn, err := net.Dial("tcp", server.ListenAddr().Host())
	t.Require().NoError(err)
	t.T().Cleanup(func() { proxyConn.Close() })

	return proxyConn
}

func makeString(size int) string {
	return strings.Repeat("abcd", size/4)
}
