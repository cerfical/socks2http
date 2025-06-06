// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	addr "github.com/cerfical/socks2http/internal/proxy/addr"

	http "net/http"

	mock "github.com/stretchr/testify/mock"

	net "net"
)

// Proxy is an autogenerated mock type for the Proxy type
type Proxy struct {
	mock.Mock
}

type Proxy_Expecter struct {
	mock *mock.Mock
}

func (_m *Proxy) EXPECT() *Proxy_Expecter {
	return &Proxy_Expecter{mock: &_m.Mock}
}

// ForwardHTTP provides a mock function with given fields: ctx, r, dstHost
func (_m *Proxy) ForwardHTTP(ctx context.Context, r *http.Request, dstHost *addr.Addr) (*http.Response, error) {
	ret := _m.Called(ctx, r, dstHost)

	if len(ret) == 0 {
		panic("no return value specified for ForwardHTTP")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, *addr.Addr) (*http.Response, error)); ok {
		return rf(ctx, r, dstHost)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, *addr.Addr) *http.Response); ok {
		r0 = rf(ctx, r, dstHost)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *http.Request, *addr.Addr) error); ok {
		r1 = rf(ctx, r, dstHost)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Proxy_ForwardHTTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForwardHTTP'
type Proxy_ForwardHTTP_Call struct {
	*mock.Call
}

// ForwardHTTP is a helper method to define mock.On call
//   - ctx context.Context
//   - r *http.Request
//   - dstHost *addr.Host
func (_e *Proxy_Expecter) ForwardHTTP(ctx interface{}, r interface{}, dstHost interface{}) *Proxy_ForwardHTTP_Call {
	return &Proxy_ForwardHTTP_Call{Call: _e.mock.On("ForwardHTTP", ctx, r, dstHost)}
}

func (_c *Proxy_ForwardHTTP_Call) Run(run func(ctx context.Context, r *http.Request, dstHost *addr.Addr)) *Proxy_ForwardHTTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*http.Request), args[2].(*addr.Addr))
	})
	return _c
}

func (_c *Proxy_ForwardHTTP_Call) Return(_a0 *http.Response, _a1 error) *Proxy_ForwardHTTP_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Proxy_ForwardHTTP_Call) RunAndReturn(run func(context.Context, *http.Request, *addr.Addr) (*http.Response, error)) *Proxy_ForwardHTTP_Call {
	_c.Call.Return(run)
	return _c
}

// OpenTunnel provides a mock function with given fields: ctx, srcConn, dstHost
func (_m *Proxy) OpenTunnel(ctx context.Context, srcConn net.Conn, dstHost *addr.Addr) (<-chan error, error) {
	ret := _m.Called(ctx, srcConn, dstHost)

	if len(ret) == 0 {
		panic("no return value specified for OpenTunnel")
	}

	var r0 <-chan error
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, net.Conn, *addr.Addr) (<-chan error, error)); ok {
		return rf(ctx, srcConn, dstHost)
	}
	if rf, ok := ret.Get(0).(func(context.Context, net.Conn, *addr.Addr) <-chan error); ok {
		r0 = rf(ctx, srcConn, dstHost)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan error)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, net.Conn, *addr.Addr) error); ok {
		r1 = rf(ctx, srcConn, dstHost)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Proxy_OpenTunnel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OpenTunnel'
type Proxy_OpenTunnel_Call struct {
	*mock.Call
}

// OpenTunnel is a helper method to define mock.On call
//   - ctx context.Context
//   - srcConn net.Conn
//   - dstHost *addr.Host
func (_e *Proxy_Expecter) OpenTunnel(ctx interface{}, srcConn interface{}, dstHost interface{}) *Proxy_OpenTunnel_Call {
	return &Proxy_OpenTunnel_Call{Call: _e.mock.On("OpenTunnel", ctx, srcConn, dstHost)}
}

func (_c *Proxy_OpenTunnel_Call) Run(run func(ctx context.Context, srcConn net.Conn, dstHost *addr.Addr)) *Proxy_OpenTunnel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(net.Conn), args[2].(*addr.Addr))
	})
	return _c
}

func (_c *Proxy_OpenTunnel_Call) Return(done <-chan error, err error) *Proxy_OpenTunnel_Call {
	_c.Call.Return(done, err)
	return _c
}

func (_c *Proxy_OpenTunnel_Call) RunAndReturn(run func(context.Context, net.Conn, *addr.Addr) (<-chan error, error)) *Proxy_OpenTunnel_Call {
	_c.Call.Return(run)
	return _c
}

// NewProxy creates a new instance of Proxy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProxy(t interface {
	mock.TestingT
	Cleanup(func())
}) *Proxy {
	mock := &Proxy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
