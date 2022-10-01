package gio

type PollEventHandler func(int, uint32) error

type PollAttachment struct {
	FD       int
	Callback PollEventHandler
}
