package wiki

import (
	"errors"
	"path/filepath"
	"sync"
	"gopkg.in/src-d/go-git.v4"
	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/wikifier"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile        string
	PrivateConfigFile string
	Opt               wikifier.PageOpt
	Auth              *authenticator.Authenticator
	pageLocks         map[string]*sync.Mutex
	_repo *git.Repository
}

// NewWiki creates a Wiki given the public and private configuration files.
func NewWiki(conf, privateConf string) (*Wiki, error) {
	conf, privateConf = filepath.FromSlash(conf), filepath.FromSlash(privateConf)
	w := &Wiki{
		ConfigFile:        conf,
		PrivateConfigFile: privateConf,
		Opt:               defaultWikiOpt,
		pageLocks:         make(map[string]*sync.Mutex),
	}

	// there's no config!
	if conf == "" {
		return nil, errors.New("no config file specified")
	}

	// guess dir.wiki from config location
	// (if the conf specifies an absolute path, this will be overwritten)
	w.Opt.Dir.Wiki = filepath.Dir(conf)

	// parse the config
	err := w.readConfig(conf)
	if err != nil {
		return nil, err
	}

	// also parse private config
	// these values override those in the public config
	if privateConf != "" {
		err := w.readConfig(privateConf)
		if err != nil {
			return nil, err
		}
	}

	// create authenticator
	w.Auth, err = authenticator.Open(filepath.Join(filepath.Dir(conf), "auth.json"))
	if err != nil {
		return nil, errors.New("init authenticator")
	}

	// no errors occurred
	return w, nil
}
