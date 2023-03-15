package gio

import (
	"net"
	"sync"
	"sync/atomic"
)

type ConnPool struct {
	sync.Mutex
	size  uint32
	cur   uint32
	conns []net.Conn
}

func newConnPool() *ConnPool {
	return &ConnPool{}
}

func (p *ConnPool) Get() net.Conn {
	return p.conns[int(atomic.AddUint32(&p.cur, 1)%p.size)]
}

func (p *ConnPool) Put(c net.Conn) {
	p.Lock()
	p.size += 1
	p.conns = append(p.conns, c)
	p.Unlock()
}

func (p *ConnPool) Puts(cs []net.Conn) {
	p.Lock()
	p.size += uint32(len(cs))
	p.conns = append(p.conns, cs...)
	p.Unlock()
}
