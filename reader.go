package gio

type Reader interface {
	Read(p []byte) (int, error)
}
