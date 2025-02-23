package wiki

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

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
	_logger       *log.Logger
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
