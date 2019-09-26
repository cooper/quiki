package wikifier

// # default options.
// our %wiki_defaults = (
//     'page.enable.title'     => 1,
//     'external.wp.name'      => 'Wikipedia',
//     'external.wp.root'      => 'http://en.wikipedia.org/wiki',
//     'external.wp.type'      => 'mediawiki',
//     'image.rounding'        => 'normal',
//     'var'                   => {} # global vars
// );

// PageOpts describes wiki/website options to a Page.
type PageOpts struct {
	Name  string // wiki name
	Dir   pageOptDir
	Root  pageOptRoot
	Image pageOptImage
}

// actual file paths
type pageOptDir struct {
	Wikifier string // path to wikifier
	Image    string // path to image directory
	Page     string // path to page directory
	Model    string // path to model directory
	Cache    string // path to cache directory
}

// web resource paths
type pageOptRoot struct {
	Wiki     string // wiki root path
	Image    string // image root path
	Category string // category root path
	Page     string // page root path
}

// image options
type pageOptImage struct {
	SizeMethod string
	Rounding   string
	Sizer      func(file string, width, height int, page *Page) (path string)
}

// defaults for Page
var defaultPageOpt = PageOpts{
	Dir: pageOptDir{
		Wikifier: ".",
		Image:    "images",
		Page:     "pages",
		Model:    "models",
		Cache:    "cache",
	},
	Root: pageOptRoot{
		Wiki:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		Page:     "/page",
	},
	Image: pageOptImage{
		SizeMethod: "javascript",
		Rounding:   "normal",
		Sizer:      defaultPageImageSizer,
	},
}

func defaultPageImageCalc(file string, width, height int, page *Page, override bool) (w, h, bigW, bigH int, fullSize bool) {
	return
}

func defaultPageImageSizer(file string, width, height int, page *Page) string {
	return ""
}
