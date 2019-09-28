package wikifier

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// # default options.
// our %wiki_defaults = (
//     'external.wp.name'      => 'Wikipedia',
//     'external.wp.root'      => 'http://en.wikipedia.org/wiki',
//     'external.wp.type'      => 'mediawiki',
//     'image.rounding'        => 'normal',
//     'var'                   => {} # global vars
// );

// PageOpt describes wiki/website options to a Page.
type PageOpt struct {
	Name     string // wiki name
	MainPage string // name of main page
	Page     PageOptPage
	Dir      PageOptDir
	Root     PageOptRoot
	Image    PageOptImage
	Category PageOptCategory
	Search   PageOptSearch
}

// PageOptPage describes option relating to a page.
type PageOptPage struct {
	EnableTitle bool // enable page title headings
	EnableCache bool // enable page caching
}

// PageOptDir describes actual filepaths to wiki resources.
type PageOptDir struct {
	Wikifier string // path to wikifier directory
	Wiki     string // path to wiki root directory
	Image    string // path to image directory
	Page     string // path to page directory
	Model    string // path to model directory
	Cache    string // path to cache directory
}

// PageOptRoot describes HTTP paths to wiki resources.
type PageOptRoot struct {
	Wiki     string // wiki root path
	Image    string // image root path
	Category string // category root path
	Page     string // page root path
}

// PageOptImage describes wiki imaging options.
type PageOptImage struct {
	Retina     []int
	SizeMethod string
	Rounding   string
	Sizer      func(file string, width, height int, page *Page) (path string)
}

// PageOptCategory describes wiki category options.
type PageOptCategory struct {
	PerPage int
}

// PageOptSearch describes wiki search options.
type PageOptSearch struct {
	Enable bool
}

// defaults for Page
var defaultPageOpt = PageOpt{
	Page: PageOptPage{
		EnableTitle: true,
		EnableCache: false,
	},
	Dir: PageOptDir{
		Wikifier: ".",
		Wiki:     "",
		Image:    "images",
		Page:     "pages",
		Model:    "models",
		Cache:    "cache",
	},
	Root: PageOptRoot{
		Wiki:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		Page:     "/page",
	},
	Image: PageOptImage{
		Retina:     []int{2, 3},
		SizeMethod: "javascript",
		Rounding:   "normal",
		Sizer:      nil,
	},
	Category: PageOptCategory{
		PerPage: 5,
	},
	Search: PageOptSearch{
		Enable: true,
	},
}

// InjectPageOpt extracts page options found in the specified page and
// injects them into the provided PageOpt pointer.
func InjectPageOpt(page *Page, opt *PageOpt) error {

	// name - wiki name
	str, err := page.GetStr("name")
	if err != nil {
		return errors.Wrap(err, "name")
	}
	if str != "" {
		opt.Name = str
	}

	// main_page - main page name
	str, err = page.GetStr("main_page")
	if err != nil {
		return errors.Wrap(err, "main_page")
	}
	if str != "" {
		opt.MainPage = str
	}

	// page.enable.title - enable page title headings
	enable, err := page.GetBool("page.enable.title")
	if err != nil {
		return errors.Wrap(err, "page.enable.title")
	}
	opt.Page.EnableTitle = enable

	// page.enable.cache - enable page caching
	enable, err = page.GetBool("page.enable.cache")
	if err != nil {
		return errors.Wrap(err, "page.enable.cache")
	}
	opt.Page.EnableCache = enable

	// image.retina - retina image scales
	if retinaStr, err := page.GetStr("image.retina"); err != nil {
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
			for _, s := range scales {
				intVal, err := strconv.Atoi(s)
				if err != nil {
					return errors.Wrap(err, "image.retina: must be list of integers")
				}
				retina = append(retina, intVal)
			}
		}
		opt.Image.Retina = retina
	}

	// image.size_method - how to determine imagebox dimensions
	str, err = page.GetStr("image.size_method")
	if err != nil {
		return errors.Wrap(err, "image.size_method")
	}
	if str != "" {
		if str != "javascript" && str != "server" {
			return errors.New("image.size_method: must be one of 'javascript' or 'server'")
		}
		opt.Image.SizeMethod = str
	}

	// cat.per_page - how many posts to show on each page of /topic
	str, err = page.GetStr("cat.per_page")
	if err != nil {
		return errors.Wrap(err, "cat.per_page")
	}
	if str != "" {
		intVal, err := strconv.Atoi(str)
		if err != nil {
			return errors.Wrap(err, "cat.per_page: must be integer")
		}
		opt.Category.PerPage = intVal
	}

	// search.enable - whether to enable search optimization
	enable, err = page.GetBool("search.enable")
	if err != nil {
		return errors.Wrap(err, "search.enable")
	}
	opt.Search.Enable = enable

	return nil
}
