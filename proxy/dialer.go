package proxy

import (
	"context"
	"net"

	"github.com/cerfical/socks2http/proxy/addr"
)

var DirectDialer Dialer = DialerFunc(directDial)

type Dialer interface {
	Dial(context.Context, *addr.Addr) (net.Conn, error)
}

type DialerFunc func(context.Context, *addr.Addr) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, h *addr.Addr) (net.Conn, error) {
	return f(ctx, h)
}

func directDial(ctx context.Context, h *addr.Addr) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", h.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
