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
		ForceGen:    false,
		Code: wikifier.PageOptCode{
			Style: "monokailight",
		},
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
		Ext:      "", // (i.e., not configured)
	},
	Image: wikifier.PageOptImage{
		Retina:         []int{2, 3},
		SizeMethod:     "server",
		Processor:      "auto",
		MaxMemoryMB:    256,
		TimeoutSeconds: 20,
		ArbitrarySizes: false, // disabled by default for security
		PregenThumbs:   "250", // default for adminifier thumbnails
		Quality:        85,
		Calc:           defaultImageCalc,
		Sizer:          defaultImageSizer,
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
		"wp": {
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

	// CRITICAL DEBUG: function called
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("CALC_CALLED: page=%s, file=%s, width=%d, height=%d", page.Name, name, width, height)
	}

	// requesting 0x0 is same as requesting full-size
	if width == 0 && height == 0 {
		return 0, 0, true
	}

	path := filepath.Join(page.Opt.Dir.Image, filepath.FromSlash(name))
	bigW, bigH := getImageDimensions(path)

	// debug dimensions
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("defaultImageCalc: original dimensions %dx%d", bigW, bigH)
	}

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

	// debug calculated dimensions
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("defaultImageCalc: calculated dimensions %dx%d", width, height)
	}

	// this must happen here to guarantee proper tracking before image pregeneration
	if w, ok := page.Wiki.(*Wiki); ok {
		img := SizedImageFromName(name)
		imageName := img.FullSizeName()

		// debug category tracking
		w.Logf("defaultImageCalc: tracking category for '%s' (from name '%s')", imageName, name)

		imageCat := w.GetSpecialCategory(imageName, CategoryTypeImage)
		imageCat.addImage(w, imageName, page, [][]int{{width, height}})
	}

	return width, height, false
}

func defaultImageSizer(name string, width, height int, page *wikifier.Page) string {
	// CRITICAL DEBUG: function called
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("SIZER_CALLED: page=%s, file=%s, width=%d, height=%d", page.Name, name, width, height)
		w.Logf("SIZER_ROOT_IMAGE: page=%s, Root.Image=%s", page.Name, page.Opt.Root.Image)
	}

	si := SizedImageFromName(name)

	// debug the parsed result
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("defaultImageSizer: parsed - Prefix='%s', RelNameNE='%s', Ext='%s'",
			si.Prefix, si.RelNameNE, si.Ext)
		if si.Prefix == "" {
			w.Logf("defaultImageSizer: WARNING - image has no prefix (root directory)")
		}
	}

	si.Width = width
	si.Height = height

	result := page.Opt.Root.Image + "/" + si.TrueName()

	// CRITICAL DEBUG: final result
	if w, ok := page.Wiki.(*Wiki); ok {
		w.Logf("SIZER_RESULT: page=%s, file=%s, TrueName=%s, final_result=%s",
			page.Name, name, si.TrueName(), result)
	}

	return result
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
		// note: page/section has been normalized and trimmed already
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
		*o.Ok = false

		warning := "Page target '" + targetName + "' does not exist"
		w.AddCheck(page.NameNE(), warning, o.Pos, func() bool {
			return w.FindPage(targetName).Exists()
		})
	}

	*o.Target = page.Opt.Root.Page + "/" + targetName + sec
	page.PageLinks[targetName] = append(page.PageLinks[targetName], o.Pos.Line)
}

func linkCategoryExists(page *wikifier.Page, o *wikifier.PageOptLinkOpts) {
	w, good := page.Wiki.(*Wiki)
	if !good {
		return
	}
	catName := wikifier.CategoryName(*o.DisplayDefault)
	*o.Ok = w.GetCategory(catName).Exists()
	if !*o.Ok {
		warning := "Category target '" + wikifier.CategoryNameNE(catName) + "' does not exist"
		w.AddCheck(page.NameNE(), warning, o.Pos, func() bool {
			return w.GetCategory(catName).Exists()
		})
	}
}
