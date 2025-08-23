package wiki

import (
	"path/filepath"

	"github.com/cooper/quiki/lock"
)

// LockPage acquires both in-memory and file-based locks for a specific page
func (w *Wiki) LockPage(pageName string) error {
	// first acquire in-memory lock
	mutex := w.GetPageLock(pageName)
	mutex.Lock()

	// find the actual page and use its lock mechanism
	page := w.FindPage(pageName)
	if page != nil {
		err := page.Lock()
		if err != nil {
			mutex.Unlock() // release in-memory lock on failure
			return err
		}
	}

	return nil
}

// UnlockPage releases both file-based and in-memory locks for a specific page
func (w *Wiki) UnlockPage(pageName string) error {
	// find the actual page and unlock it
	page := w.FindPage(pageName)
	if page != nil {
		page.Unlock()
	}

	// then release in-memory lock
	mutex := w.GetPageLock(pageName)
	mutex.Unlock()

	return nil
}

// LockWiki acquires a wiki-wide lock for bulk operations
func (w *Wiki) LockWiki() error {
	lockPath := filepath.Join(w.Opt.Dir.Wiki, "cache", "wiki.lock")
	lock := lock.New(lockPath)
	return lock.Lock()
}

// UnlockWiki releases the wiki-wide lock
func (w *Wiki) UnlockWiki() error {
	lockPath := filepath.Join(w.Opt.Dir.Wiki, "cache", "wiki.lock")
	lock := lock.New(lockPath)
	return lock.Unlock()
}

// WithWikiLock executes a function while holding the wiki-wide lock
func (w *Wiki) WithWikiLock(fn func() error) error {
	lockPath := filepath.Join(w.Opt.Dir.Wiki, "cache", "wiki.lock")
	lock := lock.New(lockPath)
	return lock.WithLock(fn)
}
