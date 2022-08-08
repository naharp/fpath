package fpath

import (
	"io"
	"os"
	"syscall"
)

func LockFile(path string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	flock := syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: io.SeekStart,
		Start:  0,
		Len:    0,
		Pid:    int32(os.Getpid()),
	}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &flock); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
