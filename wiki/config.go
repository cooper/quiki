package wiki

import (
	"fmt"

	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var defaultWikiOpt = wikifier.PageOpt{
	Page: wikifier.PageOptPage{
		EnableTitle: true,
		EnableCache: true,
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
		Calc:       defaultImageCalc,
		Sizer:      defaultImageSizer,
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
	// FIXME: where does dir.wikifier come from?
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

func defaultImageCalc(name string, width, height int, page *wikifier.Page) (finalW, finalH int) {
	path := page.Opt.Dir.Image + "/" + name
	scaleFactor := 0
	bigW, bigH := getImageDimensions(path)

	fmt.Println("path", path)

	if bigW == 0 || bigH == 0 {
		return
	}

	if width != 0 {
		scaleFactor = bigW / width
		finalW = width
		finalH = bigH / scaleFactor
		return
	}

	if height != 0 {
		scaleFactor = bigH / height
		finalW = bigW / scaleFactor
		finalH = height
	}

	return
}

func defaultImageSizer(name string, width, height int, page *wikifier.Page) string {
	si := SizedImageFromName(name)
	si.Width = width
	si.Height = height
	return page.Opt.Root.Image + "/" + si.FullName()
}
