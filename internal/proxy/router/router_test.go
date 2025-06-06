package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/router"
	"github.com/cerfical/socks2http/internal/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestRouter(t *testing.T) {
	suite.Run(t, new(RouterTest))
}

type RouterTest struct {
	suite.Suite
}

func (t *RouterTest) TestDial() {
	t.Run("routes connection requests to the destination through specified proxy", func() {
		dstAddr1 := addr.New("dst-addr-1", 80)
		proxyAddr1 := addr.New("proxy-1", 8080)

		dstAddr2 := addr.New("dst-addr-2", 80)
		proxyAddr2 := addr.New("proxy-2", 8080)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, proxyAddr1).
			Return(nil, errors.New("redirected to Proxy#1"))
		dialer.EXPECT().
			Dial(mock.Anything, proxyAddr2).
			Return(nil, errors.New("redirected to Proxy#2"))

		router := router.New(
			router.WithDialer(dialer),
			router.WithRoutes([]router.Route{{
				Hosts: []string{dstAddr1.Host},
				Proxy: router.Proxy{
					Addr:  *proxyAddr1,
					Proto: proxy.ProtoHTTP,
				},
			}, {
				Hosts: []string{dstAddr2.Host},
				Proxy: router.Proxy{
					Addr:  *proxyAddr2,
					Proto: proxy.ProtoHTTP,
				},
			}}),
		)

		// Check that the first proxy is called for the first address
		_, err := router.Dial(context.Background(), dstAddr1)
		t.ErrorContains(err, "Proxy#1")

		// Check that the second proxy is called for the second address
		_, err = router.Dial(context.Background(), dstAddr2)
		t.ErrorContains(err, "Proxy#2")
	})

	t.Run("uses default policy if routing table contains no matches", func() {
		dstAddr := addr.New("example.com", 80)
		proxyAddr := addr.New("proxy", 8081)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, proxyAddr).
			Return(nil, errors.New("redirected to Proxy"))

		router := router.New(
			router.WithDialer(dialer),
			router.WithDefaultRoute(&router.Route{
				Proxy: router.Proxy{
					Addr:  *proxyAddr,
					Proto: proxy.ProtoHTTP,
				},
			}),
		)

		_, err := router.Dial(context.Background(), dstAddr)
		t.ErrorContains(err, "Proxy")
	})
}
