package wikifier

// # default options.
// our %wiki_defaults = (
//     'external.wp.name'      => 'Wikipedia',
//     'external.wp.root'      => 'http://en.wikipedia.org/wiki',
//     'external.wp.type'      => 'mediawiki',
//     'image.rounding'        => 'normal',
//     'var'                   => {} # global vars
// );

// PageOpts describes wiki/website options to a Page.
type PageOpts struct {
	Name     string // wiki name
	Page     pageOptPage
	Dir      pageOptDir
	Root     pageOptRoot
	Image    pageOptImage
	Category pageOptCategory
	Search   pageOptSearch
}

// page options
type pageOptPage struct {
	EnableTitle bool // enable page title headings
	EnableCache bool // enable page caching
}

// actual file paths
type pageOptDir struct {
	Wikifier string // path to wikifier directory
	Wiki     string // path to wiki root directory
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
	Retina     []int
	SizeMethod string
	Rounding   string
	Sizer      func(file string, width, height int, page *Page) (path string)
}

// category options
type pageOptCategory struct {
	PerPage int
}

// search options
type pageOptSearch struct {
	Enable bool
}

// defaults for Page
var defaultPageOpt = PageOpts{
	Page: pageOptPage{
		EnableTitle: true,
		EnableCache: false,
	},
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
		Retina:     []int{2, 3},
		SizeMethod: "javascript",
		Rounding:   "normal",
		Sizer:      nil,
	},
	Category: pageOptCategory{
		PerPage: 5,
	},
	Search: pageOptSearch{
		Enable: true,
	},
}
