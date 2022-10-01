package gio

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
)

type RoundTripHandler func(conn Conn) error

type RoundTripper interface {
	RoundTrip(send, recv RoundTripHandler) error
}

type roundTripper struct {
	ctx context.Context

	opts *RTOptions

	el *eventLoop

	connPool *ConnPool
}

func NewRoundTripper(ctx context.Context, opts *RTOptions) (RoundTripper, error) {
	var err error

	rt := &roundTripper{
		ctx:      ctx,
		opts:     opts,
		connPool: newConnPool(),
	}
	rt.el, err = newEventLoop(ctx, nil, nil, &opts.Options)
	if err != nil {
		return nil, err
	}

	var (
		c  Conn
		cs []Conn
	)
	for _, endpoint := range opts.Endpoints {
		c, err = rt.Dial(endpoint.Network, endpoint.Address)
		if err != nil {

		} else {
			cs = append(cs, c)
		}
	}
	rt.connPool.Puts(cs)

	return rt, nil
}

func (rt *roundTripper) RoundTrip(send, recv RoundTripHandler) error {
	var err error

	conn := rt.connPool.Get()

	if err = send(conn); err != nil {
		return err
	}

	if err = recv(conn); err != nil {
		return err
	}

	return nil
}

func (rt *roundTripper) Dial(network, addr string) (Conn, error) {
	nc, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return rt.enroll(nc)
}

func (rt *roundTripper) enroll(nc net.Conn) (Conn, error) {
	defer nc.Close()

	sc, ok := nc.(syscall.Conn)
	if !ok {
		return nil, errors.New("failed to convert net.Conn to syscall.Conn")
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return nil, errors.New("failed to get syscall.RawConn from net.Conn")
	}

	var dupFD int
	e := rc.Control(func(fd uintptr) {
		dupFD, err = unix.Dup(int(fd))
	})
	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, e
	}

	if rt.opts.SocketSendBuffer > 0 {
		if err = SetSendBuffer(dupFD, rt.opts.SocketSendBuffer); err != nil {
			return nil, err
		}
	}
	if rt.opts.SocketRecvBuffer > 0 {
		if err = SetRecvBuffer(dupFD, rt.opts.SocketRecvBuffer); err != nil {
			return nil, err
		}
	}

	var (
		sa unix.Sockaddr
		cc *conn
	)
	switch nc.(type) {
	case *net.TCPConn:
		if rt.opts.TCPNoDelay == 1 {
			if err = SetNoDelay(dupFD, 0); err != nil {
				return nil, err
			}
		}
		if rt.opts.TCPKeepAlive > 0 {
			if err = SetKeepAlivePeriod(dupFD, int(rt.opts.TCPKeepAlive.Seconds())); err != nil {
				return nil, err
			}
		}
		if sa, _, _, _, err = GetTCPSockAddr(nc.RemoteAddr().Network(), nc.RemoteAddr().String()); err != nil {
			return nil, err
		}
		cc = newConn(rt.el, dupFD, sa, nc.LocalAddr(), nc.RemoteAddr())
	}

	//注册读写事件
	err = rt.el.onConnect(cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
