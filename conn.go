package gio

import (
	"golang.org/x/sys/unix"
	"net"
	"time"
)

type Conn interface {
	//net.Conn
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	//Flush() error
}

type conn struct {
	locker

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

func (c *conn) handleEvents(_ int, ev uint32) error {
	if ev&OutEvents != 0 {
		return c.el.onWrite(c)
	}
	if ev&InEvents != 0 {
		return c.el.onRead(c)
	}
	return nil
}

func (c *conn) Close() error {
	return c.el.closeConn(c)
}
