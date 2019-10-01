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
	Name       string // wiki name
	MainPage   string // name of main page
	Template   string // name of template
	Page       PageOptPage
	Dir        PageOptDir
	Root       PageOptRoot
	Image      PageOptImage
	Category   PageOptCategory
	Search     PageOptSearch
	Navigation []PageOptNavigation
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
	Category string // path to category directory
	Page     string // path to page directory
	Model    string // path to model directory
	Markdown string // path to markdown directory
	Cache    string // path to cache directory
}

// PageOptRoot describes HTTP paths to wiki resources.
type PageOptRoot struct {
	Wiki     string // wiki root path
	Image    string // image root path
	Category string // category root path
	Page     string // page root path
	File     string // file index path
}

// PageOptImage describes wiki imaging options.
type PageOptImage struct {
	Retina     []int
	SizeMethod string
	Rounding   string
	Calc       func(file string, width, height int, page *Page) (w, h int)
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

// PageOptNavigation represents an ordered navigation item.
type PageOptNavigation struct {
	Link    string // link
	Display string // text to display
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
		Markdown: "md",
	},
	Root: PageOptRoot{
		Wiki:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		Page:     "/page",
		File:     "",
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

	// easy string options
	pageOptString := map[string]*string{
		"name":          &opt.Name,          // wiki name
		"main_page":     &opt.MainPage,      // main page name
		"template":      &opt.Template,      // template name
		"dir.wikifier":  &opt.Dir.Wikifier,  // wikifier directory
		"dir.wiki":      &opt.Dir.Wiki,      // wiki root directory
		"dir.image":     &opt.Dir.Image,     // image directory
		"dir.page":      &opt.Dir.Page,      // page directory
		"dir.model":     &opt.Dir.Model,     // model directory
		"dir.markdown":  &opt.Dir.Markdown,  // markdown directory
		"dir.cache":     &opt.Dir.Cache,     // cache directory
		"root.wiki":     &opt.Root.Wiki,     // http path to wiki
		"root.image":    &opt.Root.Image,    // http path to images
		"root.category": &opt.Root.Category, // http path to categories
		"root.page":     &opt.Root.Page,     // http path to pages
		"root.file":     &opt.Root.File,     // http path to file index
	}
	for name, ptr := range pageOptString {
		str, err := page.GetStr(name)
		if err != nil {
			return errors.Wrap(err, name)
		}
		if str != "" {
			*ptr = str
		}
	}

	// easy bool options
	pageOptBool := map[string]*bool{
		"page.enable.title": &opt.Page.EnableTitle, // enable page title headings
		"page.enable.cache": &opt.Page.EnableCache, // enable page caching
		"search.enable":     &opt.Search.Enable,    // enable search optimization
	}
	for name, ptr := range pageOptBool {
		val, err := page.Get(name)
		if err != nil {
			return errors.Wrap(err, name)
		}
		if enable, ok := val.(bool); ok {
			*ptr = enable
		}
	}

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
	str, err := page.GetStr("image.size_method")
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

	// navigation - ordered navigation items
	obj, err := page.GetObj("navigation")
	if err != nil {
		return errors.Wrap(err, "navigation")
	}
	if obj != nil {
		navMap, ok := obj.(*Map)
		if !ok {
			return errors.New("navigation: must be map{}")
		}

		// since this runs in vars only mode,
		// have to force generate html to evalute variables
		navMap.html(page, navMap.el())

		for _, display := range navMap.OrderedKeys() {
			link, err := navMap.GetStr(display)
			display = strings.Replace(display, "_", " ", -1)
			if err != nil {
				return errors.Wrap(err, "navigation: map values must be string")
			}
			opt.Navigation = append(opt.Navigation, PageOptNavigation{
				Display: display,
				Link:    link,
			})
		}
	}

	return nil
}
