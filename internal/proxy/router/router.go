package router

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/client"
)

func New(ops ...Option) *Router {
	defaults := []Option{
		WithDefaultRoute(&Route{
			Proxy: Proxy{
				Proto: addr.ProtoDirect,
			},
		}),
		WithDialer(proxy.DirectDialer),
	}

	var r Router
	for _, op := range slices.Concat(defaults, ops) {
		op(&r)
	}
	return &r
}

func WithDefaultRoute(r *Route) Option {
	return func(rr *Router) {
		rr.defaultRoute = *r
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(r *Router) {
		r.dialer = d
	}
}

func WithRoutes(routes []Route) Option {
	return func(r *Router) {
		r.routes = routes
	}
}

type Option func(r *Router)

type Route struct {
	Hosts []string `mapstructure:"hosts"`

	Proxy Proxy `mapstructure:"proxy"`
}

type Proxy struct {
	Proto addr.Proto `mapstructure:"proto"`
	Addr  addr.Addr  `mapstructure:"addr"`
}

type Router struct {
	dialer proxy.Dialer
	routes []Route

	defaultRoute Route
}

func (r *Router) Dial(ctx context.Context, dstAddr *addr.Addr) (net.Conn, error) {
	policy := r.matchRoute(dstAddr.Host)

	client, err := client.New(
		client.WithDialer(r.dialer),
		client.WithProxyAddr(&policy.Proxy.Addr),
		client.WithProxyProto(policy.Proxy.Proto),
	)
	if err != nil {
		return nil, fmt.Errorf("new proxy client: %w", err)
	}

	return client.Dial(ctx, dstAddr)
}

func (r *Router) matchRoute(host string) *Route {
	i := slices.IndexFunc(r.routes, func(r Route) bool {
		// Check if the host matches any of the route's hosts
		for _, h := range r.Hosts {
			if strings.Contains(host, h) {
				return true
			}
		}
		return false
	})
	if i != -1 {
		return &r.routes[i]
	}
	return &r.defaultRoute
}
