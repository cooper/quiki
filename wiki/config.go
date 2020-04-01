package wiki

import (
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var defaultWikiOpt = wikifier.PageOpt{
	Page: wikifier.PageOptPage{
		EnableTitle: true,
		EnableCache: true,
	},
	Dir: wikifier.PageOptDir{
		Wiki:  "",
		Image: "images",
		Page:  "pages",
		Model: "models",
		Cache: "cache",
	},
	Root: wikifier.PageOptRoot{
		Wiki:     "", // aka /
		Page:     "", // aka /
		Image:    "/images",
		Category: "/topic",
		File:     "", // (i.e., disabled)
	},
	Image: wikifier.PageOptImage{
		Retina:     []int{2, 3},
		SizeMethod: "server",
		Calc:       defaultImageCalc,
		Sizer:      defaultImageSizer,
	},
	Category: wikifier.PageOptCategory{
		PerPage: 5,
	},
	Search: wikifier.PageOptSearch{
		Enable: true,
	},
	Link: wikifier.PageOptLink{
		ParseInternal: linkPageExists,
		ParseCategory: linkCategoryExists,
	},
	External: map[string]wikifier.PageOptExternal{
		"wp": wikifier.PageOptExternal{
			Name: "Wikipedia",
			Root: "https://en.wikipedia.org/wiki",
			Type: wikifier.PageOptExternalTypeMediaWiki},
	},
}

func (w *Wiki) readConfig(file string) error {
	// create a Page for the configuration file
	// only compute the variables
	confPage := wikifier.NewPage(file)
	confPage.VarsOnly = true

	// set this variable for use in the config
	// consider: is this needed anymore?
	confPage.Set("dir.wiki", w.Opt.Dir.Wiki)

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

func defaultImageCalc(name string, width, height int, page *wikifier.Page) (int, int, bool) {

	// requesting 0x0 is same as requesting full-size
	if width == 0 && height == 0 {
		return 0, 0, true
	}

	path := filepath.Join(page.Opt.Dir.Image, filepath.FromSlash(name))
	bigW, bigH := getImageDimensions(path)

	// original has no dimensions??
	if bigW == 0 || bigH == 0 {
		return 0, 0, true
	}

	// requesting single full-size dimension is same as requesting full-size
	if (width == bigW && height == 0) || (height == bigH && width == 0) {
		return 0, 0, true
	}

	// determine missing dimension
	width, height = calculateImageDimensions(bigW, bigH, width, height)

	// also pregenerate the image maybe
	w, ok := page.Wiki.(*Wiki)
	if ok && w.pregenerating {
		sized := SizedImageFromName(name)
		sized.Width = width
		sized.Height = height
		w.Debug("pregen image:", sized.ScaleName())
		w.DisplaySizedImageGenerate(sized, true)
	}

	return width, height, false
}

func defaultImageSizer(name string, width, height int, page *wikifier.Page) string {
	si := SizedImageFromName(name)
	si.Width = width
	si.Height = height
	return page.Opt.Root.Image + "/" + si.TrueName()
}

func linkPageExists(page *wikifier.Page, o *wikifier.PageOptLinkOpts) {
	w, good := page.Wiki.(*Wiki)
	if !good {
		return
	}

	// I don't like this
	targetName := strings.TrimPrefix(*o.Target, page.Opt.Root.Page+"/")
	sec := ""

	// remove section when checking if page exists
	if hashIdx := strings.IndexByte(targetName, '#'); hashIdx != -1 && len(targetName) >= hashIdx {
		// note: whitespace like My Page # Section has been trimmed already
		sec = "#" + targetName[hashIdx+1:]
		targetName = targetName[:hashIdx]
		if targetName == "" {
			*o.Ok = true
			return
		}
	}

	// try to find the page regardless of format/case.
	// if it exists, override name so case is correct
	targetPage := w.FindPage(targetName)
	if targetPage.Exists() {
		targetName = targetPage.NameNE()
		*o.Ok = true
	} else {
		// default behavior is lowercase, normalize
		targetName = wikifier.PageNameLink(targetName)
	}

	*o.Target = page.Opt.Root.Page + "/" + targetName + sec
	page.PageLinks[targetName] = append(page.PageLinks[targetName], 1) // FIXME: line number
}

func linkCategoryExists(page *wikifier.Page, o *wikifier.PageOptLinkOpts) {
	w, good := page.Wiki.(*Wiki)
	if !good {
		return
	}
	catName := wikifier.CategoryName(*o.DisplayDefault)
	*o.Ok = w.GetCategory(catName).Exists()
}
