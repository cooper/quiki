package wiki

import (
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var defaultWikiOpt = wikifier.PageOpt{
	Page: wikifier.PageOptPage{
		EnableTitle: true,
		EnableCache: false,
	},
	Dir: wikifier.PageOptDir{
		Wikifier: ".",
		Wiki:     "",
		Image:    "images",
		Page:     "pages",
		Model:    "models",
		Cache:    "cache",
	},
	Root: wikifier.PageOptRoot{
		Wiki:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		Page:     "/page",
	},
	Image: wikifier.PageOptImage{
		Retina:     []int{2, 3},
		SizeMethod: "server",
		Rounding:   "normal",
		Sizer:      nil, // FIXME
	},
	Category: wikifier.PageOptCategory{
		PerPage: 5,
	},
	Search: wikifier.PageOptSearch{
		Enable: true,
	},
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
	if err := wikifier.InjectPageOpt(confPage, &w.Opt); err != nil {
		return err
	}

	return nil
}
