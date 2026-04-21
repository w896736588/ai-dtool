package component

import (
	"errors"
	"os"
)

var errPlaywrightLockBusy = errors.New(`playwright lock busy`)

// TryLockFileNonBlocking 尝试对指定文件加非阻塞独占锁，成功时由调用方负责释放。
func TryLockFileNonBlocking(lockPath string) (*os.File, error) {
	return tryLockFileNonBlocking(lockPath)
}

// ReleaseLockedFile 释放文件锁并关闭文件句柄。
func ReleaseLockedFile(lockFile *os.File) error {
	if lockFile == nil {
		return nil
	}
	return releaseLockedFile(lockFile)
}

// IsPlaywrightLockBusyError 判断错误是否表示文件锁正被其他进程持有。
func IsPlaywrightLockBusyError(err error) bool {
	return errors.Is(err, errPlaywrightLockBusy)
}
