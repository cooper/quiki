package wiki

import (
	"errors"
	"path/filepath"

	"github.com/cooper/quiki/wikifier"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile        string
	PrivateConfigFile string
	Opt               wikifier.PageOpt
}

// NewWiki creates a Wiki given the public and private configuration files.
func NewWiki(conf, privateConf string) (*Wiki, error) {
	w := &Wiki{ConfigFile: conf, PrivateConfigFile: privateConf, Opt: defaultWikiOpt}

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

	// no errors occurred
	return w, nil
}
