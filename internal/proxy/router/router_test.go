package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/mocks"
	"github.com/cerfical/socks2http/internal/proxy/router"
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
		dstAddr1 := addr.NewAddr("dst-addr-1", 80)
		proxyURL1 := addr.NewURL(addr.ProtoHTTP, "proxy-1", 8080)

		dstAddr2 := addr.NewAddr("dst-addr-2", 80)
		proxyURL2 := addr.NewURL(addr.ProtoHTTP, "proxy-2", 8080)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, proxyURL1.Addr()).
			Return(nil, errors.New("redirected to Proxy#1"))
		dialer.EXPECT().
			Dial(mock.Anything, proxyURL2.Addr()).
			Return(nil, errors.New("redirected to Proxy#2"))

		router := router.New(
			router.WithDialer(dialer),
			router.WithRoutes([]router.Route{{
				Hosts: []string{dstAddr1.Host},
				Proxy: *proxyURL1,
			}, {
				Hosts: []string{dstAddr2.Host},
				Proxy: *proxyURL2,
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
		dstAddr := addr.NewAddr("example.com", 80)
		proxyURL := addr.NewURL(addr.ProtoHTTP, "proxy", 8081)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, proxyURL.Addr()).
			Return(nil, errors.New("redirected to Proxy"))

		router := router.New(
			router.WithDialer(dialer),
			router.WithDefaultRoute(&router.Route{
				Proxy: *proxyURL,
			}),
		)

		_, err := router.Dial(context.Background(), dstAddr)
		t.ErrorContains(err, "Proxy")
	})
}
