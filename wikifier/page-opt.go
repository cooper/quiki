package wikifier

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// # default options.
// our %wiki_defaults = (
//     'external.wp.name'      => 'Wikipedia',
//     'external.wp.root'      => 'http://en.wikipedia.org/wiki',
//     'external.wp.type'      => 'mediawiki',
//     'var'                   => {} # global vars
// );

// PageOpt describes wiki/website options to a Page.
type PageOpt struct {
	Name         string // wiki name
	Logo         string // logo filename, relative to image dir
	MainPage     string // name of main page
	ErrorPage    string // name of error page
	Template     string // name of template
	MainRedirect bool   // redirect on main page rather than serve root
	Page         PageOptPage
	Host         PageOptHost
	Dir          PageOptDir
	Root         PageOptRoot
	Image        PageOptImage
	Category     PageOptCategory
	Search       PageOptSearch
	Link         PageOptLink
	External     map[string]PageOptExternal
	Navigation   []PageOptNavigation
}

// PageOptPage describes option relating to a page.
type PageOptPage struct {
	EnableTitle bool        // enable page title headings
	EnableCache bool        // enable page caching
	ForceGen    bool        // force generation of page even if unchanged
	Code        PageOptCode // `code{}` block options
}

// PageOptHost describes HTTP hosts for a wiki.
type PageOptHost struct {
	Wiki string // HTTP host for the wiki
}

// PageOptCode describes options for `code{}` blocks.
type PageOptCode struct {
	Lang  string
	Style string
}

// PageOptDir describes actual filepaths to wiki resources.
type PageOptDir struct {
	Wiki     string // path to wiki root directory
	Image    string // Deprecated: path to image directory
	Category string // Deprecated: path to category directory
	Page     string // Deprecated: path to page directory
	Model    string // Deprecated: path to model directory
	Markdown string // Deprecated: path to markdown directory
	Cache    string // Deprecated: path to cache directory
}

// PageOptRoot describes HTTP paths to wiki resources.
type PageOptRoot struct {
	Wiki     string // wiki root path
	Image    string // image root path
	Category string // category root path
	Page     string // page root path
	File     string // file index path
	Ext      string // full external wiki prefix
}

// PageOptImage describes wiki imaging options.
type PageOptImage struct {
	Retina         []int
	SizeMethod     string
	MaxConcurrent  int   // max concurrent image operations (0 = auto)
	MaxMemoryMB    int64 // max memory per image in MB (0 = default 512MB)
	TimeoutSeconds int   // max processing time per image (0 = default 30s)
	ArbitrarySizes bool  // allow arbitrary image sizes not referenced in wiki content (default false)
	Calc           func(file string, width, height int, page *Page) (w, h int, fullSize bool)
	Sizer          func(file string, width, height int, page *Page) (path string)
}

// PageOptCategory describes wiki category options.
type PageOptCategory struct {
	PerPage int
}

// PageOptSearch describes wiki search options.
type PageOptSearch struct {
	Enable bool
}

// A PageOptLinkFunction sanitizes a link target.
type PageOptLinkFunction func(page *Page, opts *PageOptLinkOpts)

// PageOptLinkOpts contains options passed to a PageOptLinkFunction.
type PageOptLinkOpts struct {
	Ok             *bool   // func sets to true if the link is valid
	Target         *string // func sets to overwrite the link target
	Tooltip        *string // func sets tooltip to display
	DisplayDefault *string // func sets default text to display (if no pipe)
	*FmtOpt                // formatter options available to func
}

// PageOptLink describes functions to assist with link targets.
type PageOptLink struct {
	ParseInternal PageOptLinkFunction // internal page links
	ParseExternal PageOptLinkFunction // external wiki page links
	ParseCategory PageOptLinkFunction // category links
}

// PageOptExternalType describes
type PageOptExternalType string

const (
	// PageOptExternalTypeQuiki is the external type for quiki sites.
	PageOptExternalTypeQuiki PageOptExternalType = "quiki"

	// PageOptExternalTypeMediaWiki is the external type for MediaWiki sites.
	PageOptExternalTypeMediaWiki = "mediawiki"

	// PageOptExternalTypeNone is an external type that can be used for websites that
	// perform no normalization of page targets beyond normal URI escaping.
	PageOptExternalTypeNone = "none"
)

// PageOptExternal describes an external wiki that we can use for link targets.
type PageOptExternal struct {
	Name string              // long name (e.g. Wikipedia)
	Root string              // wiki page root (no trailing slash)
	Type PageOptExternalType // wiki type
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
		ForceGen:    false,
		Code: PageOptCode{
			Style: "monokailight",
		},
	},
	Host: PageOptHost{
		Wiki: "", // aka all hosts
	},
	Dir: PageOptDir{
		Wiki:     "",
		Image:    "images",
		Page:     "pages",
		Model:    "models",
		Cache:    "cache",
		Markdown: "md",
	},
	Root: PageOptRoot{
		Wiki:     "", // aka /
		Page:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		File:     "", // (i.e., disabled)
		Ext:      "", // (i.e., not configured)
	},
	Image: PageOptImage{
		Retina:         []int{2, 3},
		SizeMethod:     "javascript",
		ArbitrarySizes: false, // disabled by default for security
		Calc:           nil,
		Sizer:          nil,
	},
	Category: PageOptCategory{
		PerPage: 5,
	},
	Search: PageOptSearch{
		Enable: true,
	},
	Link: PageOptLink{
		ParseInternal: nil,
		ParseExternal: defaultExternalLink,
		ParseCategory: nil,
	},
	External: map[string]PageOptExternal{
		"wp": {"Wikipedia", "https://en.wikipedia.org/wiki", PageOptExternalTypeMediaWiki},
	},
}

// InjectPageOpt extracts page options found in the specified page and
// injects them into the provided PageOpt pointer.
func InjectPageOpt(page *Page, opt *PageOpt) error {

	// easy string options
	pageOptString := map[string]*string{
		"name":            &opt.Name,            // wiki name
		"logo":            &opt.Logo,            // logo filename, relative to image dir
		"main_page":       &opt.MainPage,        // main page name
		"error_page":      &opt.ErrorPage,       // error page name
		"template":        &opt.Template,        // template name
		"host.wiki":       &opt.Host.Wiki,       // wiki host
		"dir.wiki":        &opt.Dir.Wiki,        // wiki directory
		"root.wiki":       &opt.Root.Wiki,       // http path to wiki
		"root.image":      &opt.Root.Image,      // http path to images
		"root.category":   &opt.Root.Category,   // http path to categories
		"root.page":       &opt.Root.Page,       // http path to pages
		"root.file":       &opt.Root.File,       // http path to file index
		"root.ext":        &opt.Root.Ext,        // full external wiki prefix
		"page.code.lang":  &opt.Page.Code.Lang,  // code{} language
		"page.code.style": &opt.Page.Code.Style, // code{} style
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

	// these dirs are derived from wiki dir
	opt.Dir.Page = filepath.Join(opt.Dir.Wiki, "pages")
	opt.Dir.Image = filepath.Join(opt.Dir.Wiki, "images")
	opt.Dir.Model = filepath.Join(opt.Dir.Wiki, "models")
	opt.Dir.Cache = filepath.Join(opt.Dir.Wiki, "cache")
	opt.Dir.Category = filepath.Join(opt.Dir.Wiki, "cache", "category")

	// ensure they all exist
	for _, dir := range []string{
		opt.Dir.Wiki,
		opt.Dir.Page,
		opt.Dir.Image,
		opt.Dir.Model,
		opt.Dir.Cache,
		opt.Dir.Category,
	} {
		if _, err := os.Lstat(dir); err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return errors.Wrap(err, "creating "+dir)
			}
		}
	}

	// convert all HTTP roots to /
	opt.Root.Wiki = startWithSlash(filepath.ToSlash(opt.Root.Wiki))
	opt.Root.Image = startWithSlash(filepath.ToSlash(opt.Root.Image))
	opt.Root.Category = startWithSlash(filepath.ToSlash(opt.Root.Category))
	opt.Root.Page = startWithSlash(filepath.ToSlash(opt.Root.Page))
	opt.Root.File = startWithSlash(filepath.ToSlash(opt.Root.File))

	// easy bool options
	pageOptBool := map[string]*bool{
		"main_redirect":         &opt.MainRedirect,         // redirect root to main page
		"page.enable.title":     &opt.Page.EnableTitle,     // enable page title headings
		"page.enable.cache":     &opt.Page.EnableCache,     // enable page caching
		"search.enable":         &opt.Search.Enable,        // enable search optimization
		"image.arbitrary_sizes": &opt.Image.ArbitrarySizes, // allow arbitrary image sizes
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

	// TODO: External wikis

	return nil
}

func startWithSlash(s string) string {
	if s == "" || s[0] == '/' {
		return s
	}
	return "/" + s
}
