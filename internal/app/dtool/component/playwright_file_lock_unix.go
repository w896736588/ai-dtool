//go:build !windows

package component

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func tryLockFileNonBlocking(lockPath string) (*os.File, error) {
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err = unix.Flock(int(lockFile.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = lockFile.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return nil, errPlaywrightLockBusy
		}
		return nil, err
	}
	return lockFile, nil
}

func releaseLockedFile(lockFile *os.File) error {
	unlockErr := unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)
	closeErr := lockFile.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
