package gio

import (
	"golang.org/x/sys/unix"
	"net"
	"time"
)

type Conn interface {
	net.Conn
	IsActive() bool
	ID() int
}

type conn struct {
	locker

	opened bool

	el *eventLoop

	fd         int
	peer       unix.Sockaddr
	localAddr  net.Addr
	remoteAddr net.Addr

	pollAttachment *PollAttachment

	waitReadSize int64
	readTimeout  time.Duration

	inBuf  *LinkBuffer
	outBuf *LinkBuffer
}

func newConn(el *eventLoop, fd int, sa unix.Sockaddr, localAddr, remoteAddr net.Addr) *conn {
	c := &conn{
		el:         el,
		fd:         fd,
		peer:       sa,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
	c.pollAttachment = &PollAttachment{
		FD:       fd,
		Callback: c.handleEvents,
	}
	c.inBuf = NewLinkBuffer()
	c.outBuf = NewLinkBuffer()
	return c
}

func (c *conn) Read(p []byte) (int, error) {
	l := len(p)
	if l == 0 {
		return 0, nil
	}

	c.lock(read)
	defer c.unlock(read)

	buf, err := c.inBuf.Next(l)
	if err != nil {
		return 0, err
	}
	n := copy(p, buf)
	return n, nil
}

func (c *conn) Write(p []byte) (int, error) {
	c.lock(write)
	defer c.unlock(write)

	dst, _ := c.outBuf.Malloc(len(p))
	n := copy(dst, p)

	return n, nil
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) handleEvents(_ int, ev uint32) error {
	if ev&OutEvents != 0 {
		return c.el.onWrite(c)
	}
	if ev&InEvents != 0 {
		return c.el.onRead(c)
	}
	return nil
}

func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *conn) Close() error {
	//if c.el != nil {
	//
	//}
	//err := unix.Close(c.fd)
	//if err != nil {
	//	err = fmt.Errorf("failed to close fd=%d in event-loop: %v", c.fd, os.NewSyscallError("close", err))
	//	if err != nil {
	//		err = errors.New(err.Error() + " & " + err.Error())
	//	}
	//}
	//c.releaseTCP()
	//return err
	return c.el.closeConn(c)
}

func (c *conn) releaseTCP() {
	c.opened = false
	c.peer = nil
	_ = c.inBuf.Release()
	_ = c.outBuf.Release()
	c.pollAttachment = nil
}

func (c *conn) IsActive() bool {
	return c.opened
}

// ID 返回fd作为conn的id
func (c *conn) ID() int {
	return c.fd
}
