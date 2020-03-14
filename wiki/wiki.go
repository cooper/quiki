package wiki

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/cooper/go-git/v4"
	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/wikifier"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile    string
	Opt           wikifier.PageOpt
	Auth          *authenticator.Authenticator
	pageLocks     map[string]*sync.Mutex
	pregenerating bool
	_repo         *git.Repository
}

// NewWiki creates a Wiki given its directory path.
func NewWiki(path string) (*Wiki, error) {
	return NewWikiConfig(filepath.Join(path, "wiki.conf"))
}

// NewWikiConfig creates a Wiki given the configuration file path.
//
// Deprecated: Use NewWiki instead.
//
func NewWikiConfig(confPath string) (*Wiki, error) {
	confPath = filepath.FromSlash(confPath)
	w := &Wiki{
		ConfigFile: confPath,
		Opt:        defaultWikiOpt,
		pageLocks:  make(map[string]*sync.Mutex),
	}

	// there's no config!
	if confPath == "" {
		return nil, errors.New("no config file specified")
	}

	// guess dir.wiki from config location
	// (if the conf specifies an absolute path, this will be overwritten)
	w.Opt.Dir.Wiki = filepath.Dir(confPath)

	// parse the config
	err := w.readConfig(confPath)
	if err != nil {
		return nil, err
	}

	// create authenticator
	w.Auth, err = authenticator.Open(filepath.Join(filepath.Dir(confPath), "auth.json"))
	if err != nil {
		return nil, errors.New("init authenticator")
	}

	// no errors occurred
	return w, nil
}
