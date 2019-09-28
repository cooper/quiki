package wiki

import (
	"strconv"
	"strings"

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

func (w *Wiki) readConfig(file) error {

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
	if err = w.convertConfig(confPage); err != nil {
		return err
	}

	return nil
}

func (w *Wiki) convertConfig(confPage *wikifier.Page) err {
	opt := &w.Opt

	// TODO:

	// page.enable.title - enable page title headings
	enable, err := confPage.GetBool("page.enable.title")
	if err != nil {
		return errors.Wrap(err, "page.enable.title")
	}
	opt.Page.EnableTitle = enable

	// page.enable.cache - enable page caching
	enable, err = confPage.GetBool("page.enable.cache")
	if err != nil {
		return errors.Wrap(err, "page.enable.cache")
	}
	opt.Page.EnableCache = enable

	// image.retina - retina image scales
	if retinaStr, err := confPage.GetStr("image.retina"); err != nil {
		return errors.Wrap(err, "image.retina")
	} else if retinaStr != "" {
		var retina []int

		// save time - this might just be one scale
		if scale, err := strconv.Atoi(retinaStr); err == nil {
			retina = []int{scale}

		} else {
			// more than 1 scale, separated by comma

			scales := strings.Split(retinaStr, ",")
			retina = make([]int, 0, len(scales))
			i := 0
			for _, s := range scales {
				if intVal, err := strconv.Atoi(s); err != nil {
					return errors.Wrap(err, "image.retina: must be list of integers")
				}
				retina[i] = intVal
				i++
			}
		}
		opt.Image.Retina = retina
	}

	// image.size_method - how to determine imagebox dimensions
	str, err := confPage.GetStr("image.size_method")
	if err != nil {
		return errors.Wrap(err, "image.size_method")
	}
	if str != "javascript" && str != "server" {
		return errors.New("image.size_method: must be one of 'javascript' or 'server'")
	}
	opt.Image.SizeMethod = str

	// cat.per_page - how many posts to show on each page of /topic
	str, err = confPage.GetStr("cat.per_page")
	if err != nil {
		return errors.Wrap(err, "cat.per_page")
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		return errors.Wrap(err, "cat.per_page: must be integer")
	}
	opt.Category.PerPage = intVal

	// 'search.enable' - whether to enable search optimization
	enable, err = confPage.Get("search.enable")
	if err != nil {
		return errors.Wrap(err, "search.enable")
	}
	opt.Search.Enable = enable

}
