package wiki

import (
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var defaultWikiOpt = wikifier.PageOpts{
	// TODO:
	// 'page.enable.cache'             => 1,
	// 'image.enable.restriction'      => 1,
	// 'image.enable.cache'            => 1,
	// 'image.enable.retina'           => 3,
	// 'image.enable.tracking'         => 1,
	// 'image.enable.pregeneration'    => 1,
	// 'image.rounding'                => 'up',
	// 'image.size_method'             => 'server',
	// 'image.sizer'                   => \&_wiki_default_sizer,   # from Images
	// 'image.calc'                    => \&_wiki_default_calc,    # from Images
	// 'cat.per_page'                  => 5,
	// 'search.enable'                 => 1
}

func (w *Wiki) readConfig() error {

	// there's no config!
	if w.ConfigFile == "" {
		return errors.New("no config file specified")
	}

	// create a Page for the configuration file
	// only compute the variables
	confPage := wikifier.NewPage(w.ConfigFile)
	confPage.VarsOnly = true
	w.conf = confPage

	// set these variables for use in the config
	confPage.Set("dir.wiki", w.Opt.Dir.Wiki)
	confPage.Set("dir.wikifier", w.Opt.Dir.Wikifier)

	// parse the config
	if err := confPage.Parse(); err != nil {
		return errors.Wrap(err, "failed to parse configuration")
	}

	// TODO: Extract global wiki variables

	// private configuration
	if w.PrivateConfigFile != "" {

		// create a Page for the private configuration file
		// only compute the variables
		pconfPage := wikifier.NewPage(w.PrivateConfigFile)
		pconfPage.VarsOnly = true
		w.pconf = pconfPage

		// set these variables for use in the private config
		pconfPage.Set("dir.wiki", w.Opt.Dir.Wiki)
		pconfPage.Set("dir.wikifier", w.Opt.Dir.Wikifier)

		// parse the private config
		if err := pconfPage.Parse(); err != nil {
			return errors.Wrap(err, "failed to parse private configuration")
		}

	} else {
		// if there is no private config, assume the main one also contains
		// private wiki settings

		w.pconf = confPage
	}

	return nil
}
