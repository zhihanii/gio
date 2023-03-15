package gio

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type eventLoop struct {
	ctx context.Context

	opts *Options

	poller *Poller

	buf []byte

	connections map[int]*conn

	eh EventHandler

	failCallback PollEventHandler
}

func newEventLoop(ctx context.Context, eh EventHandler, cb PollEventHandler, opts *Options) (*eventLoop, error) {
	var err error
	el := &eventLoop{
		ctx:          ctx,
		opts:         opts,
		buf:          make([]byte, opts.ReadBufferCap),
		connections:  make(map[int]*conn),
		eh:           eh,
		failCallback: cb,
	}
	el.poller, err = openPoller()
	if err != nil {
		return nil, err
	}
	return el, nil
}

func (el *eventLoop) activeReactor() {
	err := el.poller.Poll(func(fd int, e uint32) error {
		if c, ok := el.connections[fd]; ok {
			if e&OutEvents != 0 {
				return el.onWrite(c)
			}
			if e&InEvents != 0 {
				return el.onRead(c)
			}
			return nil
		}
		if el.failCallback != nil {
			return el.failCallback(fd, e)
		}
		return nil
	})
	if err != nil {
		//log
	}
}

func (el *eventLoop) onConnect(c *conn) error {
	var err error
	//注册读事件
	if err = el.poller.AddReadWrite(c.pollAttachment); err != nil {
		return err
	}
	c.opened = true
	el.connections[c.fd] = c
	return el.eh.OnConnect(el.ctx, c)
}

//func (el *eventLoop) open(c *conn) error {
//	//注册写事件
//	return el.poller.AddWrite(c.pollAttachment)
//}

func (el *eventLoop) onRead(c *conn) error {
	n, err := unix.Read(c.fd, el.buf)
	if err != nil || n == 0 {
		if err == unix.EAGAIN {
			return nil
		}
		if n == 0 {
			err = unix.ECONNRESET
		}
		//return close conn
	}

	_, _ = c.Write(el.buf[:n])

	_ = el.eh.OnRead(el.ctx, c)

	return nil
}

func (el *eventLoop) onWrite(c *conn) error {
	iov := c.outBuf.Peek()
	if len(iov) == 0 {
		return nil
	}
	var (
		n   int
		err error
	)
	if len(iov) > 1 {
		n, err = unix.Writev(c.fd, iov)
	} else {
		n, err = unix.Write(c.fd, iov[0])
	}
	c.outBuf.Discard(n)
	if err != nil {
		if err == unix.EAGAIN {
			err = nil
		} else {
			//close conn
		}
	}
	return err
}

func (el *eventLoop) closeConn(c *conn) error {
	var (
		n   int
		err error
	)
	iov := c.outBuf.Peek()
	if len(iov) > 0 {
		n, err = unix.Writev(c.fd, iov)
		if err != nil {
			//log
		} else {
			c.outBuf.Discard(n)
		}
	}

	err0, err1 := el.poller.Delete(c.fd), unix.Close(c.fd)
	if err0 != nil {
		err = fmt.Errorf("failed to delete fd=%d from poller in event-loop: %v", c.fd, err0)
	}
	if err1 != nil {
		err1 = fmt.Errorf("failed to close fd=%d in event-loop: %v", c.fd, os.NewSyscallError("close", err1))
		if err != nil {
			err = errors.New(err.Error() + " & " + err1.Error())
		} else {
			err = err1
		}
	}

	delete(el.connections, c.fd)
	el.eh.OnClose(c, err)
	c.releaseTCP()

	return err
}
