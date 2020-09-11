package util

import (
	"net"
	"time"
)

// TimeoutConn net.Conn with Read/Write timeout. from https://qiita.com/kwi/items/b38d6273624ad3f6ae79
type TimeoutConn struct {
	net.Conn
	timeout time.Duration
}

// NewTimeoutConn create timeout conn
func NewTimeoutConn(conn net.Conn, timeout time.Duration) *TimeoutConn {
	return &TimeoutConn{
		Conn:    conn,
		timeout: timeout,
	}
}

func (c *TimeoutConn) Read(p []byte) (n int, err error) {
	if c.timeout > 0 {
		if err := c.Conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}

	return c.Conn.Read(p)
}

func (c *TimeoutConn) Write(p []byte) (n int, err error) {
	if c.timeout > 0 {
		if err := c.Conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}

	return c.Conn.Write(p)
}
