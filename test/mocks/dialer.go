// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	addr "github.com/cerfical/socks2http/addr"

	mock "github.com/stretchr/testify/mock"

	net "net"
)

// Dialer is an autogenerated mock type for the Dialer type
type Dialer struct {
	mock.Mock
}

type Dialer_Expecter struct {
	mock *mock.Mock
}

func (_m *Dialer) EXPECT() *Dialer_Expecter {
	return &Dialer_Expecter{mock: &_m.Mock}
}

// Dial provides a mock function with given fields: _a0, _a1
func (_m *Dialer) Dial(_a0 context.Context, _a1 *addr.Host) (net.Conn, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Dial")
	}

	var r0 net.Conn
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *addr.Host) (net.Conn, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *addr.Host) net.Conn); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *addr.Host) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Dialer_Dial_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Dial'
type Dialer_Dial_Call struct {
	*mock.Call
}

// Dial is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *addr.Host
func (_e *Dialer_Expecter) Dial(_a0 interface{}, _a1 interface{}) *Dialer_Dial_Call {
	return &Dialer_Dial_Call{Call: _e.mock.On("Dial", _a0, _a1)}
}

func (_c *Dialer_Dial_Call) Run(run func(_a0 context.Context, _a1 *addr.Host)) *Dialer_Dial_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*addr.Host))
	})
	return _c
}

func (_c *Dialer_Dial_Call) Return(_a0 net.Conn, _a1 error) *Dialer_Dial_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Dialer_Dial_Call) RunAndReturn(run func(context.Context, *addr.Host) (net.Conn, error)) *Dialer_Dial_Call {
	_c.Call.Return(run)
	return _c
}

// NewDialer creates a new instance of Dialer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDialer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Dialer {
	mock := &Dialer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
