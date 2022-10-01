package gio

import "time"

type Options struct {
	ReadBufferCap    int
	SocketSendBuffer int
	SocketRecvBuffer int
	TCPKeepAlive     time.Duration
	TCPNoDelay       int
}

type RTOptions struct {
	Endpoints []NetAddr
	Options
}

type NetAddr struct {
	Network string
	Address string
}
