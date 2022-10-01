package gio

import "golang.org/x/sys/unix"

const (
	InitListSize = 128
	MaxListSize  = 1024

	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	OutEvents = ErrEvents | unix.EPOLLOUT
	InEvents  = ErrEvents | unix.EPOLLIN
)

type epollEvent = unix.EpollEvent

type eventList struct {
	size   int
	events []epollEvent
}

func newEventList(size int) *eventList {
	return &eventList{
		size:   size,
		events: make([]epollEvent, size),
	}
}
