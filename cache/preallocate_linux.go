package cache

import (
	"os"
	"syscall"
)

func preallocate(f *os.File, size int64) error {
	if size <= 0 {
		return nil
	}
	err := syscall.Fallocate(int(f.Fd()), 0, 0, size)
	if errNo, ok := err.(syscall.Errno); ok && errNo == syscall.EOPNOTSUPP {
		return syscall.Ftruncate(int(f.Fd()), size)
	}
	return err
}
