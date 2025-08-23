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

// PregenerateManager defines the interface we need for page generation
type PregenerateManager interface {
	GeneratePageSync(string, bool) any
}

// Manager coordinates file system monitoring across all wikis.
type Manager struct {
	mu       sync.RWMutex
	watchers map[string]*wikiWatcher
	ctx      context.Context
	cancel   context.CancelFunc
}

// wikiWatcher represents monitoring for a single wiki.
type wikiWatcher struct {
	wiki               *wiki.Wiki
	pregenerateManager PregenerateManager // pregeneration manager for unified queue system
	watcher            *fsnotify.Watcher
	ctx                context.Context
	cancel             context.CancelFunc
	symlinkTargets     sync.Map // maps target file paths to slice of symlink relative paths
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

// AddWikiWithPregeneration starts monitoring a wiki for file changes with pregeneration manager.
func (m *Manager) AddWikiWithPregeneration(w *wiki.Wiki, pregenerateManager PregenerateManager) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	wikiName := w.Name()

	// stop existing watcher if present
	if existing, exists := m.watchers[wikiName]; exists {
		existing.stop()
		delete(m.watchers, wikiName)
	}

	// create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(m.ctx)
	ww := &wikiWatcher{
		wiki:               w,
		pregenerateManager: pregenerateManager,
		watcher:            watcher,
		ctx:                ctx,
		cancel:             cancel,
	}

	// add directories to watch
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

			// if this is in the page directory and is a symlink, track it and watch the target
			if dir == w.Opt.Dir.Page && info.Mode()&os.ModeSymlink != 0 {
				if target, err := filepath.EvalSymlinks(path); err == nil {
					// get relative path of the symlink
					if relPath, err := filepath.Rel(w.Opt.Dir.Page, path); err == nil {
						// track this symlink target
						ww.addSymlinkTarget(target, relPath)

						// add target to watcher if it's not already watched
						if err := watcher.Add(target); err == nil {
							log.Printf("Watching symlink target: %s -> %s", relPath, target)
						}
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Warning: failed to watch directory %s: %v", dir, err)
		}
	}

	m.watchers[wikiName] = ww

	// start monitoring in background
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

// addSymlinkTarget tracks a symlink target for regeneration
func (ww *wikiWatcher) addSymlinkTarget(targetPath, symlinkRelPath string) {
	if existing, ok := ww.symlinkTargets.Load(targetPath); ok {
		if symlinks, ok := existing.([]string); ok {
			// add to existing slice if not already present
			for _, existing := range symlinks {
				if existing == symlinkRelPath {
					return
				}
			}
			ww.symlinkTargets.Store(targetPath, append(symlinks, symlinkRelPath))
		}
	} else {
		ww.symlinkTargets.Store(targetPath, []string{symlinkRelPath})
	}
}

// removeSymlinkTarget removes a symlink from target tracking
func (ww *wikiWatcher) removeSymlinkTarget(targetPath, symlinkRelPath string) {
	if existing, ok := ww.symlinkTargets.Load(targetPath); ok {
		if symlinks, ok := existing.([]string); ok {
			// remove from slice
			for i, symlink := range symlinks {
				if symlink == symlinkRelPath {
					newSymlinks := append(symlinks[:i], symlinks[i+1:]...)
					if len(newSymlinks) == 0 {
						// no more symlinks for this target, remove entirely
						ww.symlinkTargets.Delete(targetPath)
						// stop watching the target since nothing points to it now
						_ = ww.watcher.Remove(targetPath)
					} else {
						ww.symlinkTargets.Store(targetPath, newSymlinks)
					}
					return
				}
			}
		}
	}
}

// getSymlinkTargets returns all symlinks that point to the given target path.
func (ww *wikiWatcher) getSymlinkTargets(targetPath string) []string {
	if symlinks, ok := ww.symlinkTargets.Load(targetPath); ok {
		if slice, ok := symlinks.([]string); ok {
			return slice
		}
	}
	return nil
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
		return
	}

	// check if this is a symlink target file that we should regenerate symlinks for
	if symlinks := ww.getSymlinkTargets(abs); symlinks != nil {
		for _, symlinkRelPath := range symlinks {
			log.Printf("Symlink target changed, regenerating symlink: %s -> %s", symlinkRelPath, abs)
			ww.regeneratePage(symlinkRelPath)
		}
		return
	}

	// check if this might be a broken symlink target (file deleted)
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		// clean up any symlink targets that pointed to this file
		ww.symlinkTargets.Range(func(key, value interface{}) bool {
			targetPath := key.(string)
			if targetPath == abs {
				if symlinks, ok := value.([]string); ok {
					for _, symlinkRelPath := range symlinks {
						log.Printf("Symlink target deleted, regenerating broken symlink page: %s (target %s)", symlinkRelPath, abs)
						ww.regeneratePage(symlinkRelPath)
					}
				}
				// remove from tracking and stop watching the target path
				ww.symlinkTargets.Delete(targetPath)
				_ = ww.watcher.Remove(targetPath)
			}
			return true
		})
	}
}

// handlePageEvent processes page file changes
func (ww *wikiWatcher) handlePageEvent(relPath string, event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create, fsnotify.Write:
		abs := filepath.Join(ww.wiki.Opt.Dir.Page, relPath)

		// if this is a newly created symlink, track its target
		if fi, err := os.Lstat(abs); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			if target, err := filepath.EvalSymlinks(abs); err == nil {
				ww.addSymlinkTarget(target, relPath)
				// also add target to watcher
				if err := ww.watcher.Add(target); err == nil {
					log.Printf("Watching new symlink target: %s -> %s", relPath, target)
				}
			}
		} else if event.Op == fsnotify.Write {
			// check if this is an existing symlink whose target changed
			ww.handlePotentialSymlinkChange(relPath)
		}

		ww.regeneratePage(relPath)

	case fsnotify.Rename, fsnotify.Remove:
		log.Printf("Page removed: %s", relPath)

		// if this was a symlink, clean up target tracking
		abs := filepath.Join(ww.wiki.Opt.Dir.Page, relPath)
		// try to get the target before it's gone (for renames, this might still work)
		if target, err := os.Readlink(abs); err == nil {
			// resolve to absolute path
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(abs), target)
			}
			if resolvedTarget, err := filepath.Abs(target); err == nil {
				ww.removeSymlinkTarget(resolvedTarget, relPath)
			}
		}

		// clean up cache
		cacheDir := ww.wiki.Opt.Dir.Cache
		os.Remove(filepath.Join(cacheDir, "page", relPath+".cache"))
		os.Remove(filepath.Join(cacheDir, "page", relPath+".txt"))
	}
}

// handlePotentialSymlinkChange checks if a symlink target has changed
func (ww *wikiWatcher) handlePotentialSymlinkChange(relPath string) {
	abs := filepath.Join(ww.wiki.Opt.Dir.Page, relPath)
	if fi, err := os.Lstat(abs); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if newTarget, err := filepath.EvalSymlinks(abs); err == nil {
			// find if this symlink was previously tracked under a different target
			var oldTarget string
			ww.symlinkTargets.Range(func(key, value interface{}) bool {
				targetPath := key.(string)
				if symlinks, ok := value.([]string); ok {
					for _, symlink := range symlinks {
						if symlink == relPath && targetPath != newTarget {
							oldTarget = targetPath
							return false // stop iteration
						}
					}
				}
				return true
			})

			if oldTarget != "" {
				// remove from old target and add to new target
				ww.removeSymlinkTarget(oldTarget, relPath)
				ww.addSymlinkTarget(newTarget, relPath)
				// add new target to watcher
				if err := ww.watcher.Add(newTarget); err == nil {
					log.Printf("Symlink target changed: %s -> %s (was %s)", relPath, newTarget, oldTarget)
				}
				// if old target has no more symlinks, stop watching it
				if _, ok := ww.symlinkTargets.Load(oldTarget); !ok {
					_ = ww.watcher.Remove(oldTarget)
				}
			}
		}
	}
}

// regenerates a single page with proper locking
func (ww *wikiWatcher) regeneratePage(relPath string) {
	log.Printf("Page changed, regenerating: %s", relPath)
	ww.pregenerateManager.GeneratePageSync(relPath, true)
}
