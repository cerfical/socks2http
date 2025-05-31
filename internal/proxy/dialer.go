package proxy

import (
	"context"
	"net"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

var DirectDialer Dialer = DialerFunc(func(ctx context.Context, a *addr.Addr) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "tcp", a.String())
})

type Dialer interface {
	Dial(context.Context, *addr.Addr) (net.Conn, error)
}

type DialerFunc func(context.Context, *addr.Addr) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, h *addr.Addr) (net.Conn, error) {
	return f(ctx, h)
}
