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

func (w *Wiki) readConfig(file string) error {

	// create a Page for the configuration file
	// only compute the variables
	confPage := wikifier.NewPage(file)
	confPage.VarsOnly = true

	// set these variables for use in the config
	// FIXME: where do these come from?
	confPage.Set("dir.wiki", w.Opt.Dir.Wiki)
	confPage.Set("dir.wikifier", w.Opt.Dir.Wikifier)

	// parse the config
	if err := confPage.Parse(); err != nil {
		return errors.Wrap(err, "failed to parse configuration "+file)
	}

	// convert the config to wikifier.PageOpt
	if err := wikifier.InjectPageOpts(confPage, &w.Opt); err != nil {
		return err
	}

	return nil
}
