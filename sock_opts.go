package gio

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func SetKeepAlivePeriod(fd, secs int) error {
	if secs <= 0 {
		return errors.New("invalid time duration")
	}
	if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)); err != nil {
		return err
	}
	if err := os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs)); err != nil {
		return err
	}
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs))
}

func SetSendBuffer(fd, size int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, size)
}

func SetRecvBuffer(fd, size int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, size)
}

func SetNoDelay(fd, noDelay int) error {
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, noDelay))
}
