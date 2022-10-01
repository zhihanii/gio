package gio

type Writer interface {
	Write(p []byte) (int, error)
	//Flush() error
}
