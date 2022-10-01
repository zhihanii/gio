package gio

import (
	"encoding/binary"
	"io"
)

type ReadWriter interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
}

type ByteBuffer struct {
	rw  ReadWriter
	buf [16]byte
}

func NewByteBuffer(rw ReadWriter) *ByteBuffer {
	return &ByteBuffer{rw: rw}
}

func (b *ByteBuffer) ReadUint32() (uint32, error) {
	_, err := b.rw.Read(b.buf[:4])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b.buf[:4]), nil
}

func (b *ByteBuffer) ReadBytes(data []byte) (int, error) {
	return b.rw.Read(data)
}

func (b *ByteBuffer) WriteUint16(n uint16) (int, error) {
	binary.BigEndian.PutUint16(b.buf[:2], n)
	return b.rw.Write(b.buf[:2])
}

func (b *ByteBuffer) WriteUint32(n uint32) (int, error) {
	binary.BigEndian.PutUint32(b.buf[:4], n)
	return b.rw.Write(b.buf[:4])
}

func (b *ByteBuffer) WriteString(s string) (int, error) {
	n, err := b.WriteUint16(uint16(len(s)))
	if err != nil {
		return n, err
	}
	return io.WriteString(b.rw, s)
}

func (b *ByteBuffer) WriteBytes(data []byte) (int, error) {
	n, err := b.WriteUint32(uint32(len(data)))
	if err != nil {
		return n, err
	}
	return b.rw.Write(data)
}

//func (b *ByteBuffer) Flush() error {
//	return b.w.Flush()
//}
