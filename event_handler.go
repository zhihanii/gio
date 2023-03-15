package gio

import "context"

type EventHandler interface {
	OnConnect(ctx context.Context, conn Conn) error
	OnRead(ctx context.Context, conn Conn) error
	OnClose(conn Conn, err error)
}
