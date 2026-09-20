//go:build unix

package filelock

import (
	"os"

	"golang.org/x/sys/unix"
)

func exclusive(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX)
}

func unlock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
