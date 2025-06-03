package stubs

import (
	"net"
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
