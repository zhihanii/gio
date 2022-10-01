package gio

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"time"
)

type Listener interface {
	Serve(ln net.Listener) error
}

type tcpListener struct {
	ctx context.Context

	opts *Options

	el *eventLoop

	fd int
	ln *net.TCPListener

	stop chan error
}

func NewTCPListener(ctx context.Context, eh EventHandler, opts *Options) (Listener, error) {
	var err error
	l := &tcpListener{
		ctx:  ctx,
		opts: opts,
		stop: make(chan error),
	}
	l.el, err = newEventLoop(ctx, eh, l.accept, opts)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *tcpListener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l *tcpListener) Serve(ln net.Listener) error {
	var (
		file *os.File
		err  error
	)

	tcpLn, ok := ln.(*net.TCPListener)
	if !ok {
		return errors.New("")
	}
	file, err = tcpLn.File()
	if err != nil {
		return err
	}
	l.fd = int(file.Fd())
	l.ln = tcpLn

	err = l.el.poller.AddRead(&PollAttachment{
		FD:       l.fd,
		Callback: l.accept,
	})
	if err != nil {
		return err
	}

	go l.el.activeReactor()

	return err
}

func (l *tcpListener) accept(_ int, _ uint32) error {
	//返回新的fd, 用于conn的读写
	nfd, sa, err := unix.Accept(l.fd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return os.NewSyscallError("accept", err)
	}
	if err = os.NewSyscallError("failed to set nonblock", unix.SetNonblock(nfd, true)); err != nil {
		return err
	}

	remoteAddr := SockAddrToTCPOrUnixAddr(sa)
	if l.opts.TCPKeepAlive > 0 {
		err = SetKeepAlivePeriod(nfd, int(l.opts.TCPKeepAlive/time.Second))
		//log
	}

	c := newConn(l.el, nfd, sa, l.Addr(), remoteAddr)
	return l.el.onConnect(c)
}

func (l *tcpListener) Stop() {
	l.quit(nil)
}

func (l *tcpListener) waitStop() error {
	return <-l.stop
}

func (l *tcpListener) quit(err error) {
	select {
	case l.stop <- err:
	default:
	}
}
