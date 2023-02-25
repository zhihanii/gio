package gio

import (
	"golang.org/x/sys/unix"
	"os"
	"runtime"
)

type Poller struct {
	fd  int
	efd int
}

func openPoller() (*Poller, error) {
	var err error
	p := new(Poller)
	if p.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		return nil, err
	}
	if p.efd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC); err != nil {
		_ = p.Close()
		return nil, err
	}
	if err = p.AddRead(&PollAttachment{FD: p.efd}); err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}

func (p *Poller) Poll(callback func(fd int, e uint32) error) error {
	el := newEventList(InitListSize)

	msec := -1
	for {
		//等待网络事件
		n, err := unix.EpollWait(p.fd, el.events, msec)
		if n == 0 || (n < 0 && err == unix.EINTR) {
			msec = -1
			runtime.Gosched()
			continue
		} else if err != nil {
			//log
			return err
		}
		msec = 0

		for i := 0; i < n; i++ {
			e := &el.events[i]
			if fd := int(e.Fd); fd != p.efd {
				err = callback(fd, e.Events)
				if err != nil {

				}
			} else {

			}
		}
	}
}

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

func (p *Poller) AddReadWrite(pa *PollAttachment) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, pa.FD, &unix.EpollEvent{Fd: int32(pa.FD), Events: readWriteEvents}))
}

func (p *Poller) AddRead(pa *PollAttachment) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, pa.FD, &unix.EpollEvent{Fd: int32(pa.FD), Events: readEvents}))
}

func (p *Poller) AddWrite(pa *PollAttachment) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, pa.FD, &unix.EpollEvent{Fd: int32(pa.FD), Events: writeEvents}))
}

func (p *Poller) Delete(fd int) error {
	return os.NewSyscallError("epoll_ctl del", unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil))
}

func (p *Poller) Close() error {
	if err := os.NewSyscallError("close", unix.Close(p.fd)); err != nil {
		return err
	}
	return os.NewSyscallError("close", unix.Close(p.efd))
}
