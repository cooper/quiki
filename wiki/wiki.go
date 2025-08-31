package wiki

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/wikifier"
	"github.com/go-git/go-git/v5"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile     string
	Opt            wikifier.PageOpt
	Auth           *authenticator.Authenticator
	pageLocks      map[string]*sync.Mutex
	pageLocksmu    sync.RWMutex
	imageLocks     map[string]*sync.Mutex // locks for image generation
	imageLocksmu   sync.RWMutex
	pregenerating  bool
	checks         []Check
	checkMu        sync.Mutex
	currentBatcher *categoryBatcher // current batching context, if any
	_repo          *git.Repository
	_logger        *log.Logger
}

// NewWiki creates a Wiki given its directory path.
func NewWiki(path string) (*Wiki, error) {

	if path == "" {
		return nil, errors.New("no wiki path specified")
	}

	confPath := filepath.Join(path, "wiki.conf")
	w := &Wiki{
		ConfigFile: confPath,
		Opt:        defaultWikiOpt,
		pageLocks:  make(map[string]*sync.Mutex),
		imageLocks: make(map[string]*sync.Mutex),
	}

	w.Opt.Dir.Wiki = path

	// parse the config
	err := w.readConfig(confPath)
	if err != nil {
		return nil, err
	}

	// create authenticator
	w.Auth, err = authenticator.Open(filepath.Join(filepath.Dir(confPath), "auth.json"))
	if err != nil {
		return nil, errors.Wrap(err, "init authenticator")
	}
	if w.Auth.IsNew {
		w.Log("created wiki authentication file")
	}

	// no errors occurred
	return w, nil
}

// Name returns the wiki's name based on its directory.
func (w *Wiki) Name() string {
	return filepath.Base(w.Opt.Dir.Wiki)
}

// GetPageLock returns the mutex for a specific page, creating it if necessary.
func (w *Wiki) GetPageLock(pageName string) *sync.Mutex {
	w.pageLocksmu.Lock()
	defer w.pageLocksmu.Unlock()

	if _, exists := w.pageLocks[pageName]; !exists {
		w.pageLocks[pageName] = new(sync.Mutex)
	}
	return w.pageLocks[pageName]
}

// GetImageLock returns the mutex for a specific image, creating it if necessary.
func (w *Wiki) GetImageLock(imageName string) *sync.Mutex {
	w.imageLocksmu.Lock()
	defer w.imageLocksmu.Unlock()

	if _, exists := w.imageLocks[imageName]; !exists {
		w.imageLocks[imageName] = new(sync.Mutex)
	}
	return w.imageLocks[imageName]
}
