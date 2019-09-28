package wiki

import (
	"errors"

	"github.com/cooper/quiki/wikifier"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile        string
	PrivateConfigFile string
	Opt               wikifier.PageOpts
}

// NewWiki creates a Wiki given the public and private configuration files.
func NewWiki(conf, privateConf string) (*Wiki, error) {
	w := &Wiki{ConfigFile: conf, PrivateConfigFile: privateConf, Opt: defaultWikiOpt}

	// there's no config!
	if conf == "" {
		return nil, errors.New("no config file specified")
	}

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

// NewPage creates a Page given its filepath and configures it for
// use with this Wiki.
func (w *Wiki) NewPage(filePath string) *wikifier.Page {
	p := wikifier.NewPage(filePath)
	p.Opt = w.Opt
	return p
}
