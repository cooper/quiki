// Package monitor provides a file monitor that pre-generates
// wiki pages and images each time a change is detected on the filesystem.
package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/wiki"
	"github.com/fsnotify/fsnotify"
)

type wikiMonitor struct {
	w        *wiki.Wiki
	watcher  *fsnotify.Watcher
	watching map[string]bool
}

// WatchWiki starts a file monitor loop for the provided wiki.
func WatchWiki(w *wiki.Wiki) {

	// creates a new file watcher
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	// create monitor
	mon := wikiMonitor{w, watcher, make(map[string]bool)}

	// find all the directories
	walkDir := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// since fsnotify can watch all the files in a directory, watchers only need
		// to be added to each nested directory
		if fi.Mode().IsDir() {
			abs, _ := filepath.Abs(path)
			mon.watching[abs] = true
			return watcher.Add(abs)
		}

		return nil
	}

	roots := map[string]func(mon wikiMonitor, event fsnotify.Event, abs string){
		w.Opt.Dir.Page:     handlePageEvent,
		w.Opt.Dir.Image:    handleImageEvent,
		w.Opt.Dir.Model:    handleModelEvent,
		w.Opt.Dir.Markdown: handleMarkdownEvent,
	}

	// watch each of the content roots
	for root, handler := range roots {
		delete(roots, root)
		root, _ = filepath.Abs(root)
		if root == "" {
			continue
		}
		roots[root] = handler
		if err := filepath.Walk(root, walkDir); err != nil {
			fmt.Println("ERROR", err)
		}
	}

	//
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:

				// don't waste any time with these
				if event.Op == fsnotify.Chmod {
					continue
				}

				// find absolute path
				abs, err := filepath.Abs(event.Name)
				if err != nil {
					log.Println(err)
					continue
				}

				// new directory created -- add to monitor
				fi, err := os.Lstat(abs)
				if err == nil && fi.IsDir() && event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("adding dir", abs)
					mon.watching[abs] = true
					watcher.Add(abs)
					continue
				}

				// a directory we were watching has been deleted
				// note: this catches renames also
				if mon.watching[abs] && event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("DELETE DIR", event)
					delete(mon.watching, abs)
					// watcher.Remove(abs) // nvm, it does this automatically
					continue
				}

				// file change; pass it on to handlers
				for root, handler := range roots {
					if strings.HasPrefix(abs, root+"/") {
						handler(mon, event, abs)
						break
					}
				}

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done
}

func handlePageEvent(mon wikiMonitor, event fsnotify.Event, abs string) {
	dirPage, _ := filepath.Abs(mon.w.Opt.Dir.Page)
	name := strings.TrimPrefix(abs, dirPage+"/")

	switch event.Op {

	case fsnotify.Create, fsnotify.Write:

		// this is a symlink; ignore it
		// FIXME: only skip if the target is also in the page root?
		if fi, err := os.Lstat(abs); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return
		}

		mon.w.DisplayPageDraft(name, true)

	case fsnotify.Rename, fsnotify.Remove:
		// TODO: w.PurgePage() or similar
		os.Remove(mon.w.Opt.Dir.Cache + "/page/" + name + ".cache")
		os.Remove(mon.w.Opt.Dir.Cache + "/page/" + name + ".txt")

	}
}

func handleImageEvent(mon wikiMonitor, event fsnotify.Event, abs string) {}

func handleModelEvent(mon wikiMonitor, event fsnotify.Event, abs string) {}

func handleMarkdownEvent(mon wikiMonitor, event fsnotify.Event, abs string) {}
