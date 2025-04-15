package proxserv_test

import (
	"context"
	"testing"
	"time"

	"github.com/cerfical/socks2http/proxy/proxserv"
	"github.com/cerfical/socks2http/test/stubs"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTest))
}

type ServerTest struct {
	suite.Suite
}

func (t *ServerTest) TestServe() {
	t.Run("performs graceful shutdown on context cancellation", func() {
		listener := stubs.NewIdleListener(1000, 50*time.Millisecond)

		server, err := proxserv.New()
		t.Require().NoError(err)

		serveCtx, serveStop := context.WithCancel(context.Background())
		serveErr := make(chan error)
		go func() {
			serveErr <- server.Serve(serveCtx, listener)
		}()

		// Stop the server and check that all previously open connections are now closed
		serveStop()
		t.Require().NoError(<-serveErr)
		t.Equal(0, listener.OpenConns())
	})
}
