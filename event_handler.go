package gio

import "context"

type EventHandler interface {
	OnRead(ctx context.Context, conn Conn) error
}
