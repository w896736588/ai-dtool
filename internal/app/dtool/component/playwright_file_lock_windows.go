//go:build windows

package component

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func tryLockFileNonBlocking(lockPath string) (*os.File, error) {
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	var overlapped windows.Overlapped
	err = windows.LockFileEx(
		windows.Handle(lockFile.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&overlapped,
	)
	if err != nil {
		_ = lockFile.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, errPlaywrightLockBusy
		}
		return nil, err
	}
	return lockFile, nil
}

func releaseLockedFile(lockFile *os.File) error {
	var overlapped windows.Overlapped
	unlockErr := windows.UnlockFileEx(windows.Handle(lockFile.Fd()), 0, 1, 0, &overlapped)
	closeErr := lockFile.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
