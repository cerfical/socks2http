package router

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/proxcli"
)

func New(ops ...Option) *Router {
	defaults := []Option{
		WithDefaultPolicy(&Policy{
			Proxy: ProxyConfig{
				Proto: proxy.ProtoDirect,
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

func WithDefaultPolicy(p *Policy) Option {
	return func(r *Router) {
		r.defaultRoute = *p
	}
}

func WithDialer(d proxy.Dialer) Option {
	return func(r *Router) {
		r.dialer = d
	}
}

type Option func(r *Router)

type Policy struct {
	Proxy ProxyConfig
}

type ProxyConfig struct {
	Proto proxy.Proto
	Addr  addr.Addr
}

type RouteTable map[string]Policy

type Router struct {
	dialer proxy.Dialer
	routes RouteTable

	defaultRoute Policy
}

func (r *Router) Route(host string, p *Policy) *Router {
	if r.routes == nil {
		r.routes = make(RouteTable)
	}
	r.routes[host] = *p
	return r
}

func (r *Router) Dial(ctx context.Context, dstAddr *addr.Addr) (net.Conn, error) {
	policy := r.matchRoute(dstAddr.Host)

	client, err := proxcli.New(
		proxcli.WithDialer(r.dialer),
		proxcli.WithProxyAddr(&policy.Proxy.Addr),
		proxcli.WithProxyProto(policy.Proxy.Proto),
	)
	if err != nil {
		return nil, fmt.Errorf("new proxy client: %w", err)
	}

	return client.Dial(ctx, dstAddr)
}

func (r *Router) matchRoute(host string) *Policy {
	if policy, ok := r.routes[host]; ok {
		return &policy
	}
	return &r.defaultRoute
}
