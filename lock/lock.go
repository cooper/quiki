package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Lock represents a file-based lock
type Lock struct {
	path string
	file *os.File
}

// New creates a new file lock at the specified path
func New(path string) *Lock {
	return &Lock{path: path}
}

// Lock acquires the file lock with a timeout
func (l *Lock) Lock() error {
	return l.LockWithTimeout(30 * time.Second)
}

// LockWithTimeout acquires the file lock with a specified timeout
func (l *Lock) LockWithTimeout(timeout time.Duration) error {
	// ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(l.path), 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %v", err)
	}

	start := time.Now()
	for {
		// try to create the lock file exclusively
		file, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			l.file = file
			// write process info to the lock file
			fmt.Fprintf(file, "pid:%d\ntime:%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
			file.Sync()
			return nil
		}

		// if we can't create the file, check if it's stale
		if l.isStale() {
			os.Remove(l.path) // remove stale lock
			continue
		}

		// check timeout
		if time.Since(start) > timeout {
			return fmt.Errorf("failed to acquire lock %s: timeout after %v", l.path, timeout)
		}

		// wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}
}

// Unlock releases the file lock
func (l *Lock) Unlock() error {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	return os.Remove(l.path)
}

// isStale checks if the lock file is stale (process no longer exists)
func (l *Lock) isStale() bool {
	// for now, just check if the file is older than 5 minutes
	// in a more sophisticated implementation, we could check if the PID is still running
	info, err := os.Stat(l.path)
	if err != nil {
		return true // if we can't stat it, consider it stale
	}
	return time.Since(info.ModTime()) > 5*time.Minute
}

// WithLock executes a function while holding the lock
func (l *Lock) WithLock(fn func() error) error {
	if err := l.Lock(); err != nil {
		return err
	}
	defer l.Unlock()
	return fn()
}

// WithLockTimeout executes a function while holding the lock with a timeout
func (l *Lock) WithLockTimeout(timeout time.Duration, fn func() error) error {
	if err := l.LockWithTimeout(timeout); err != nil {
		return err
	}
	defer l.Unlock()
	return fn()
}
