package proxcli_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/proxy/proxcli"
	"github.com/stretchr/testify/suite"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTest))
}

type ClientTest struct {
	suite.Suite
}

func (t *ClientTest) TestDial() {
	t.Run("establishes a direct connection to a server if Direct protocol is used", func() {
		dstHost := addr.NewHost("localhost", 8080)
		client, err := proxcli.New(
			proxcli.WithProxyAddr(addr.New(addr.Direct, "", 0)),
			proxcli.WithDialer(proxy.DialerFunc(
				func(ctx context.Context, h *addr.Host) (net.Conn, error) {
					if *h != *dstHost {
						return nil, errors.New("expected a direct connection to the host")
					}
					return nil, nil
				},
			)),
		)
		t.Require().NoError(err)

		errChan := make(chan error)
		go func() {
			_, err := client.Dial(context.Background(), dstHost)
			errChan <- err
		}()

		t.Require().NoError(<-errChan)
	})
}
