package proxy

import (
	"context"
	"net"

	"github.com/cerfical/socks2http/addr"
)

type Dialer interface {
	Dial(context.Context, *addr.Host) (net.Conn, error)
}

type DialerFunc func(context.Context, *addr.Host) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	return f(ctx, h)
}
