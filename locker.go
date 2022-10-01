package gio

import "sync/atomic"

type key int32

const (
	closing key = iota
	processing
	flushing
	finalizing

	read
	write

	total
)

type locker struct {
	keychain [total]int32
}

func (l *locker) lock(k key) bool {
	for {
		if atomic.CompareAndSwapInt32(&l.keychain[k], 0, 1) {
			break
		}
	}
	return true
}

func (l *locker) unlock(k key) {
	atomic.StoreInt32(&l.keychain[k], 0)
}
