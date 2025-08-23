// Package monitor provides file system monitoring for wiki changes.
package monitor

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cooper/quiki/wiki"
	"github.com/fsnotify/fsnotify"
)

// Manager coordinates file system monitoring across all wikis.
type Manager struct {
	mu       sync.RWMutex
	watchers map[string]*wikiWatcher
	ctx      context.Context
	cancel   context.CancelFunc
}

// wikiWatcher represents monitoring for a single wiki.
type wikiWatcher struct {
	wiki    *wiki.Wiki
	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

// Global manager instance
var globalManager *Manager
var managerOnce sync.Once

// GetManager returns the global monitor manager instance.
func GetManager() *Manager {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		globalManager = &Manager{
			watchers: make(map[string]*wikiWatcher),
			ctx:      ctx,
			cancel:   cancel,
		}
	})
	return globalManager
}

// AddWiki starts monitoring a wiki for file changes.
func (m *Manager) AddWiki(w *wiki.Wiki) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	wikiName := w.Name()

	// Stop existing watcher if present
	if existing, exists := m.watchers[wikiName]; exists {
		existing.stop()
		delete(m.watchers, wikiName)
	}

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(m.ctx)
	ww := &wikiWatcher{
		wiki:    w,
		watcher: watcher,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Add directories to watch
	dirs := []string{
		w.Opt.Dir.Page,
		w.Opt.Dir.Image,
		w.Opt.Dir.Model,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking
			}
			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})

		if err != nil {
			log.Printf("Warning: failed to watch directory %s: %v", dir, err)
		}
	}

	m.watchers[wikiName] = ww

	// Start monitoring in background
	go ww.monitor()

	log.Printf("Started monitoring wiki: %s", wikiName)
	return nil
}

// RemoveWiki stops monitoring a wiki.
func (m *Manager) RemoveWiki(wikiName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ww, exists := m.watchers[wikiName]; exists {
		ww.stop()
		delete(m.watchers, wikiName)
		log.Printf("Stopped monitoring wiki: %s", wikiName)
	}
}

// Stop gracefully shuts down all monitoring.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, ww := range m.watchers {
		ww.stop()
		delete(m.watchers, name)
	}

	m.cancel()
	log.Println("Monitor manager stopped")
}

// stop gracefully stops this wiki watcher.
func (ww *wikiWatcher) stop() {
	ww.cancel()
	if ww.watcher != nil {
		ww.watcher.Close()
	}
}

// monitor is the main monitoring loop for a wiki.
func (ww *wikiWatcher) monitor() {
	defer ww.stop()

	for {
		select {
		case <-ww.ctx.Done():
			return

		case event := <-ww.watcher.Events:
			ww.handleEvent(event)

		case err := <-ww.watcher.Errors:
			if err != nil {
				log.Printf("Watcher error for wiki %s: %v", ww.wiki.Name(), err)
			}
		}
	}
}

// handleEvent processes a file system event.
func (ww *wikiWatcher) handleEvent(event fsnotify.Event) {
	// Skip chmod events
	if event.Op == fsnotify.Chmod {
		return
	}

	// Small delay to debounce rapid changes
	time.Sleep(50 * time.Millisecond)

	abs, err := filepath.Abs(event.Name)
	if err != nil {
		return
	}

	// Handle directory creation
	if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
		if event.Op&fsnotify.Create == fsnotify.Create {
			ww.watcher.Add(abs)
			log.Printf("Added directory to watch: %s", abs)
		}
		return
	}

	// Determine if this is a page file
	pageDir, _ := filepath.Abs(ww.wiki.Opt.Dir.Page)
	if relPath, err := filepath.Rel(pageDir, abs); err == nil && !filepath.IsAbs(relPath) {
		ww.handlePageEvent(relPath, event)
	}
}

// handlePageEvent processes page file changes.
func (ww *wikiWatcher) handlePageEvent(relPath string, event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create, fsnotify.Write:
		// Skip if it's a symlink
		abs := filepath.Join(ww.wiki.Opt.Dir.Page, relPath)
		if fi, err := os.Lstat(abs); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return
		}

		log.Printf("Page changed, regenerating: %s", relPath)

		// Use page locking to coordinate with other operations
		pageLock := ww.wiki.GetPageLock(relPath)
		pageLock.Lock()
		defer pageLock.Unlock()

		// Force regeneration
		ww.wiki.DisplayPageDraft(relPath, true)

	case fsnotify.Rename, fsnotify.Remove:
		log.Printf("Page removed: %s", relPath)

		// Clean up cache
		cacheDir := ww.wiki.Opt.Dir.Cache
		os.Remove(filepath.Join(cacheDir, "page", relPath+".cache"))
		os.Remove(filepath.Join(cacheDir, "page", relPath+".txt"))
	}
}
