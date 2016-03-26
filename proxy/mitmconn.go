package proxy

import (
	"io"
	"net"
	"net/http"
	"time"
)

type mitmConn struct {
	w          FlushWriter
	r          io.Reader
	remoteAddr addr
	closed     chan struct{}
}

type FlushWriter interface {
	io.Writer
	http.Flusher
}

func newMitmConn(w FlushWriter, r io.Reader, remoteAddr string) *mitmConn {
	return &mitmConn{
		w:          w,
		r:          r,
		remoteAddr: addr(remoteAddr),
		closed:     make(chan struct{}),
	}
}

func (c *mitmConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *mitmConn) Write(b []byte) (int, error) {
	n, err := c.w.Write(b)
	c.w.Flush()
	return n, err
}

func (c *mitmConn) Close() error {
	close(c.closed)
	return nil
}

func (c *mitmConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *mitmConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (c *mitmConn) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mitmConn) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mitmConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}

type addr string

func (a addr) String() string {
	return string(a)
}

func (a addr) Network() string {
	return "tcp"
}
