package proxserv_test

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
	"github.com/cerfical/socks2http/proxy/proxserv"
	"github.com/cerfical/socks2http/proxy/proxtest"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/suite"
)

const (
	megabyte = 1 << 20

	numInRowRequests = 10
	maxPayloadSize   = megabyte * 10
)

var unreachableHost = addr.NewHost("0.0.0.0", 1234)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTest))
}

type ServerTest struct {
	suite.Suite
}

func (t *ServerTest) TestNew() {
	tests := map[string]struct {
		options proxserv.Option
		want    func(*proxserv.Server)
		err     func(error)
	}{
		"uses http-localhost-8080 as the default listen address": {
			want: func(s *proxserv.Server) {
				t.Equal(addr.New(addr.HTTP, "localhost", 8080), s.ListenAddr())
			},
		},

		"uses a non-default listen address if one is provided": {
			options: proxserv.WithListenAddr(addr.New(addr.HTTP, "localhost", 8181)),
			want: func(s *proxserv.Server) {
				t.Equal(addr.New(addr.HTTP, "localhost", 8181), s.ListenAddr())
			},
		},

		"rejects unsupported protocol schemes": {
			options: proxserv.WithListenAddr(addr.New("SOCKS", "", 0)),
			err: func(err error) {
				t.ErrorContains(err, "SOCKS")
			},
		},

		"starts to listen on the specified address": {
			options: proxserv.WithListenAddr(addr.New(addr.HTTP, "localhost", 0)),
			want: func(s *proxserv.Server) {
				t.assertHostIsReachable(&s.ListenAddr().Host)
			},
		},

		"allocates a listen port if one was not provided": {
			options: proxserv.WithListenAddr(addr.New(addr.HTTP, "localhost", 0)),
			want: func(s *proxserv.Server) {
				t.NotZero(s.ListenAddr().Host.Port)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			ops := []proxserv.Option{}
			if test.options != nil {
				ops = append(ops, test.options)
			}

			serv, err := proxserv.New(context.Background(), ops...)
			if test.err != nil {
				test.err(err)
			} else {
				t.T().Cleanup(func() { serv.Stop() })

				t.Require().NoError(err)
				test.want(serv)
			}
		})
	}
}

func (t *ServerTest) TestServe_HTTP() {
	serverHost := proxtest.StartHTTPEchoServer(t.T())
	tests := map[string]struct {
		want func(net.Conn)
	}{
		"supports forwarding of a single HTTP request": {
			want: func(proxyConn net.Conn) {
				t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, "everything is OK", serverHost, proxyConn)
			},
		},

		"supports forwarding of a single HTTP request with a large payload": {
			want: func(proxyConn net.Conn) {
				largeMsg := makeString(maxPayloadSize)
				t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, largeMsg, serverHost, proxyConn)
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
			proxyConn := t.openProxyConn(addr.HTTP)
			test.want(proxyConn)
		})
	}
}

func (t *ServerTest) TestServe_HTTPS() {
	serverHost := proxtest.StartHTTPSEchoServer(t.T())
	tests := map[string]struct {
		want func(net.Conn)
	}{
		"supports tunneling of HTTPS requests": {
			want: func(proxyConn net.Conn) {
				t.assertHTTPEchoesBack(http.StatusOK, http.MethodConnect, "", serverHost, proxyConn)

				proxyConn = newTLSConn(proxyConn)
				for i := range numInRowRequests {
					msg := fmt.Sprintf("message #%v", i+1)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, msg, serverHost, proxyConn)
				}
			},
		},

		"supports tunneling of HTTPS requests with large payloads": {
			want: func(proxyConn net.Conn) {
				t.assertHTTPEchoesBack(http.StatusOK, http.MethodConnect, "", serverHost, proxyConn)

				proxyConn = newTLSConn(proxyConn)
				for range numInRowRequests {
					largeMsg := makeString(maxPayloadSize)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, largeMsg, serverHost, proxyConn)
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			proxyConn := t.openProxyConn(addr.HTTP)
			test.want(proxyConn)
		})
	}
}

func (t *ServerTest) TestServe_SOCKS() {
	serverHost := proxtest.StartHTTPEchoServer(t.T())
	tests := map[string]struct {
		want func(net.Conn)
	}{
		"supports tunneling of HTTP requests": {
			want: func(proxyConn net.Conn) {
				req := socks.NewRequest(socks.V4, socks.Connect, serverHost)
				reply := t.roundTripSOCKS(req, proxyConn)
				t.Require().Equal(socks.Granted, reply)

				for range numInRowRequests {
					msg := makeString(maxPayloadSize)
					t.assertHTTPEchoesBack(http.StatusOK, http.MethodPost, msg, serverHost, proxyConn)
				}
			},
		},

		"responds with a Request Rejected if the server is unreachable": {
			want: func(proxyConn net.Conn) {
				req := socks.NewRequest(socks.V4, socks.Connect, unreachableHost)
				reply := t.roundTripSOCKS(req, proxyConn)
				t.Require().Equal(socks.Rejected, reply)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			proxyConn := t.openProxyConn(addr.SOCKS4)
			test.want(proxyConn)
		})
	}
}

func (t *ServerTest) assertHostIsReachable(h *addr.Host) {
	t.T().Helper()

	conn, err := net.Dial("tcp", h.String())
	t.NoError(err)
	t.T().Cleanup(func() { conn.Close() })
}

func (t *ServerTest) assertHTTPEchoesBack(status int, method, msg string, h *addr.Host, c net.Conn) {
	t.T().Helper()

	req := t.newHTTPRequest(method, msg, h)
	resp := t.roundTripHTTP(req, c)

	t.Equal(status, resp.StatusCode)
	t.Equal(msg, t.readString(resp.Body))
}

func (t *ServerTest) newHTTPRequest(method, body string, h *addr.Host) *http.Request {
	t.T().Helper()

	r, err := http.NewRequest(method, "", strings.NewReader(body))
	t.Require().NoError(err)
	r.Host = h.String()

	return r
}

func (t *ServerTest) roundTripHTTP(r *http.Request, serverConn net.Conn) *http.Response {
	t.T().Helper()

	t.Require().NoError(r.Write(serverConn))

	resp, err := http.ReadResponse(bufio.NewReader(serverConn), r)
	t.Require().NoError(err)

	return resp
}

func (t *ServerTest) roundTripSOCKS(r *socks.Request, serverConn net.Conn) socks.Reply {
	t.T().Helper()

	t.Require().NoError(r.Write(serverConn))
	reply, err := socks.ReadReply(bufio.NewReader(serverConn))
	t.Require().NoError(err)

	return reply
}

func (t *ServerTest) readString(r io.Reader) string {
	t.T().Helper()

	bytes, err := io.ReadAll(r)
	t.Require().NoError(err)
	return string(bytes)
}

func (t *ServerTest) openProxyConn(proto string) (proxyConn net.Conn) {
	t.T().Helper()

	proxyHost := proxtest.StartProxyServer(t.T(), proto)
	proxyConn, err := net.Dial("tcp", proxyHost.String())
	t.Require().NoError(err)
	t.T().Cleanup(func() { proxyConn.Close() })

	return proxyConn
}

func newTLSConn(c net.Conn) net.Conn {
	return tls.Client(c, &tls.Config{InsecureSkipVerify: true})
}

func makeString(size int) string {
	return strings.Repeat("abcd", size/4)
}
