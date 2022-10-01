package gio

const (
	MaxSize = 1024 * 4
)

type LinkBuffer struct {
	head  *linkBufferNode
	read  *linkBufferNode
	write *linkBufferNode
}

func NewLinkBuffer() *LinkBuffer {
	node := newLinkBufferNode(MaxSize)
	return &LinkBuffer{
		head:  node,
		read:  node,
		write: node,
	}
}

func (b *LinkBuffer) Next(n int) ([]byte, error) {
	if n <= b.read.Len() {
		return b.read.Next(n), nil
	}
	var pIdx int
	p := make([]byte, n)
	var l int
	for ack := n; ack > 0; ack -= l {
		l = b.read.Len()
		if l >= ack {
			pIdx += copy(p[pIdx:], b.read.Next(ack))
			break
		} else if l > 0 {
			pIdx += copy(p[pIdx:], b.read.Next(l))
		}
		if b.read.IsFull() {
			b.read = b.read.next
		}
	}
	return p, nil
}

func (b *LinkBuffer) Malloc(n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	b.growth(n)
	return b.write.Malloc(n), nil
}

func (b *LinkBuffer) Peek() [][]byte {
	var iov [][]byte
	for node := b.write; node != nil; node = node.next {
		if node.Len() > 0 {
			iov = append(iov, node.Peek(-1))
		}
	}
	return iov
}

func (b *LinkBuffer) Discard(n int) {
	for n > 0 {
		if n >= b.write.Len() {
			n -= b.write.Len()
			b.write.off = b.write.malloc
			if b.write.IsFull() {
				b.write = b.write.next
			} else {
				break
			}
		} else {
			b.write.off += n
			n = 0
		}
	}
}

func (b *LinkBuffer) Release() error {
	for b.head != b.read && b.head != b.write {
		node := b.head
		b.head = b.head.next
		node.Release()
	}
	return nil
}

func (b *LinkBuffer) growth(n int) {
	if n <= 0 {
		return
	}
	if cap(b.write.buf)-b.write.malloc < n {
		b.write.next = newLinkBufferNode(n)
		b.write = b.write.next
	}
}

type linkBufferNode struct {
	buf    []byte
	off    int
	malloc int
	prev   *linkBufferNode
	next   *linkBufferNode
}

func newLinkBufferNode(size int) *linkBufferNode {
	return &linkBufferNode{
		buf:    make([]byte, size),
		off:    0,
		malloc: 0,
		prev:   nil,
		next:   nil,
	}
}

func (b *linkBufferNode) Len() int {
	return b.malloc - b.off
}

func (b *linkBufferNode) IsFull() bool {
	return cap(b.buf) == b.malloc
}

func (b *linkBufferNode) Malloc(n int) []byte {
	malloc := b.malloc
	b.malloc += n
	return b.buf[malloc:b.malloc]
}

func (b *linkBufferNode) Next(n int) []byte {
	off := b.off
	b.off += n
	return b.buf[off:b.off]
}

func (b *linkBufferNode) Peek(n int) []byte {
	if n == -1 {
		return b.buf[b.off:b.malloc]
	}
	return b.buf[b.off : b.off+n]
}

func (b *linkBufferNode) Release() {
	b.buf, b.prev, b.next = nil, nil, nil
}
