package server_test

import (
	"net"
	"sync/atomic"
	"time"
)

// NewDummyConn creates a new [DummyConn].
func NewDummyConn() *DummyConn {
	return &DummyConn{}
}

// DummyConn reads and writes no data.
type DummyConn struct{}

func (*DummyConn) Read([]byte) (int, error)  { return 0, nil }
func (*DummyConn) Write([]byte) (int, error) { return 0, nil }

func (*DummyConn) Close() error { return nil }

func (*DummyConn) LocalAddr() net.Addr  { return nil }
func (*DummyConn) RemoteAddr() net.Addr { return nil }

func (*DummyConn) SetDeadline(time.Time) error      { return nil }
func (*DummyConn) SetReadDeadline(time.Time) error  { return nil }
func (*DummyConn) SetWriteDeadline(time.Time) error { return nil }

// NewIdleListener creates a new [IdleListener] with the specified number of active connections and time delay.
func NewIdleListener(numConns int, idleTime time.Duration) *IdleListener {
	return &IdleListener{
		idleTime: idleTime,
		numConns: numConns,

		closed: make(chan struct{}),
	}
}

// IdleListener opens a number of idle connections and then closes them one by one after a delay.
type IdleListener struct {
	idleTime time.Duration
	numConns int

	numOpenConns atomic.Int64

	closed chan struct{}
}

func (l *IdleListener) Accept() (net.Conn, error) {
	// Open dummy connections until the limit is reached
	if l.numOpenConns.Add(1) > int64(l.numConns) {
		l.numOpenConns.Add(-1)

		// After that, the listener can be closed
		close(l.closed)
		return nil, net.ErrClosed
	}

	c1, c2 := net.Pipe()
	go func() {
		<-l.closed

		// After the listener is closed, hang the connection for a while
		time.AfterFunc(l.idleTime, func() {
			c2.Close()
		})
	}()

	// Track the number of closed connections
	return &connCloser{c1, func() {
		l.numOpenConns.Add(-1)
	}}, nil
}

func (l *IdleListener) Close() error {
	<-l.closed
	return nil
}

func (l *IdleListener) Addr() net.Addr {
	return nil
}

func (l *IdleListener) OpenConns() int {
	return int(l.numOpenConns.Load())
}

type connCloser struct {
	net.Conn

	close func()
}

func (t *connCloser) Close() error {
	t.close()
	return t.Conn.Close()
}
