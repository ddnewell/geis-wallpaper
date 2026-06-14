//go:build darwin || linux

package platform

import (
	"os"
	"syscall"
)

// FileLock is an advisory exclusive lock held on an open file for its lifetime.
type FileLock struct {
	f *os.File
}

// TryLock attempts a non-blocking exclusive flock at path. It returns (nil, nil)
// — no error, no lock — if another process already holds it, so the caller can
// simply skip this cycle (prevents overlapping launchd runs from racing on the
// frame ring).
func TryLock(path string) (*FileLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, nil
		}
		return nil, err
	}
	return &FileLock{f: f}, nil
}

// Unlock releases the lock and closes the file.
func (l *FileLock) Unlock() error {
	if l == nil || l.f == nil {
		return nil
	}
	if err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN); err != nil {
		l.f.Close()
		return err
	}
	return l.f.Close()
}
