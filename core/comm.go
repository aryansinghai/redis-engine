package core

import "syscall"

type FDComm struct {
	Fd int
}

func (c FDComm) Read(p []byte) (n int, err error) {
	return syscall.Read(c.Fd, p)
}

func (c FDComm) Write(p []byte) (n int, err error) {
	return syscall.Write(c.Fd, p)
}
