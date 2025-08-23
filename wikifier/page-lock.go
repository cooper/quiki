package wikifier

import (
	"path/filepath"

	"github.com/cooper/quiki/lock"
)

// Lock acquires a file-based lock for standalone page operations
func (p *Page) Lock() error {
	if p.FilePath == "" {
		return nil
	}

	dir := filepath.Dir(p.FilePath)
	name := filepath.Base(p.FilePath)
	lockPath := filepath.Join(dir, ".lock."+name)

	lock := lock.New(lockPath)
	return lock.Lock()
}

// Unlock releases the file-based lock for standalone page operations
func (p *Page) Unlock() error {
	if p.FilePath == "" {
		return nil
	}

	dir := filepath.Dir(p.FilePath)
	name := filepath.Base(p.FilePath)
	lockPath := filepath.Join(dir, ".lock."+name)

	lock := lock.New(lockPath)
	return lock.Unlock()
}

// WithLock executes a function while holding a file-based lock on the page
func (p *Page) WithLock(fn func() error) error {
	if p.FilePath == "" {
		return fn()
	}

	dir := filepath.Dir(p.FilePath)
	name := filepath.Base(p.FilePath)
	lockPath := filepath.Join(dir, ".lock."+name)

	lock := lock.New(lockPath)
	return lock.WithLock(fn)
}
