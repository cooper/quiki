package wikifier

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var externalImageRegex = regexp.MustCompile(`^(.+)://`)

type imageBlock struct {
	file, path, alt, link, lastName string
	align, float, author, license   string
	width, height                   int
	parseFailed, useJS              bool
	parsedDimensions                bool
	fullSize                        bool
	scales                          []int
	*Map
}

type imagebox struct {
	*imageBlock
}

// new image{}
func newImageBlock(name string, b *parserBlock) block {
	return &imageBlock{Map: newMapBlock("", b).(*Map)}
}

// new imagebox{}
func newImagebox(name string, b *parserBlock) block {
	return &imagebox{newImageBlock(name, b).(*imageBlock)}
}

// image{} or imagebox{} parse
func (image *imageBlock) parse(page *Page) {
	image.Map.parse(page)

	// fetch string values from map
	image.file = image.getString("file")
	image.alt = image.getString("alt")
	image.link = image.getString("link")
	image.align = image.getString("align")
	image.float = image.getString("float")
	image.author = image.getString("author")
	image.license = image.getString("license")

	// fetch dimensions
	if !image.parsedDimensions {
		image.width = image.getPx("width")
		image.height = image.getPx("height")
		image.parsedDimensions = true
	}

	// compatibility
	if image.align == "" {
		image.align = image.getString("float")
	}

	// determine alt text
	if image.alt == "" {
		image.alt = image.file
	}

	// no dimensions. if it's an infobox we can guess it
	if image.width == 0 && image.height == 0 && image.parentBlock().blockType() == "infobox" {
		image.width = 270
	}

	// no file - this is mandatory
	if image.file == "" {
		image.warn(image.getKeyPos("file"), "No file specified for image")
		image.parseFailed = true
		return
	}

	image.path = image.file
	_, image.lastName = filepath.Split(image.file)

	// ##############
	// ### SIZING ###
	// ##############

	sizeMethod := strings.ToLower(page.Opt.Image.SizeMethod)

	if externalImageRegex.MatchString(image.file) {
		// if the file is an absolute URL, we cannot size the image
		// do nothing

	} else if sizeMethod == "javascript" {
		// use javascript image sizing
		//
		// uses full-size images directly and uses javascript to size imageboxes
		// - voids the validity as XHTML 1.0 Strict
		// - causes slight flash on page load when images are scaled

		// inject javascript resizer if no width is given
		if image.width == 0 {
			image.useJS = true
			image.width = 200 // dummy will be overridden by javascript
		}

		// path is file relative to image root (full size image)
		image.path = page.Opt.Root.Image + "/" + image.file

	} else if sizeMethod == "server" {
		// use server-size image sizing
		//
		// - maintains XHTML 1.0 Strict validity
		// - eliminates flash on page load
		// - faster page load (since image files are smaller)
		// - require read access to local image directory

		// these must be provided by wiki
		if page.Opt.Image.Sizer == nil || page.Opt.Image.Calc == nil {
			image.warn(image.openPos, "image.sizer and image.calc required with image.size_method 'server'")
			image.parseFailed = true
			return
		}

		// determine dimensions
		var calcWidth, calcHeight int
		calcWidth, calcHeight, image.fullSize = page.Opt.Image.Calc(
			image.file,
			image.width,
			image.height,
			page,
		)

		// path is as returned by the function that sizes the image
		image.path = page.Opt.Image.Sizer(
			image.file,
			calcWidth,
			calcHeight,
			page,
		)

		// remember that the page uses this image in these dimensions
		// consider: should we remember the retina scales? I guess it doesn't really DEPEND on them
		page.Images[image.file] = append(page.Images[image.file], []int{calcWidth, calcHeight})

		// for each retina scale, determine whether the scaled dimensions would exceed full-size
		for _, scale := range page.Opt.Image.Retina {
			_, _, tooBig := page.Opt.Image.Calc(
				image.file,
				scale*image.width,
				scale*image.height,
				page,
			)
			if !tooBig {
				image.scales = append(image.scales, scale)
			}
		}

	} else {
		// note: this should never happen because the config parser validates it
		image.warn(image.openPos, "image.size_method neither 'javascript' nor 'server'")
		image.parseFailed = true
		return
	}
}

// image{} html
func (image *imageBlock) html(page *Page, el element) {
	image.imageHTML(false, page, el)
}

// imagebox{} html
func (image *imagebox) html(page *Page, el element) {
	image.imageHTML(true, page, el)
}

// image{} or imagebox{} html
func (image *imageBlock) imageHTML(isBox bool, page *Page, el element) {
	image.Map.html(page, el)

	// image parse failed, so no need to waste any time here
	if image.parseFailed {
		return
	}

	// add the appropriate float class
	if isBox {
		if image.float == "" {
			image.float = "right"
		}
		el.addClass("imagebox-" + image.float)
	} else if image.float != "" {
		el.addClass("image-" + image.float)
	}

	url, _ := url.Parse(image.path)
	isAbsolute := url != nil && url.IsAbs()

	// retina--
	//
	// skip is using full size image (since it can't be scaled any larger than that)
	// skip if the image URL is absolute (not an image served by this wiki)
	//
	srcset := ""
	if !image.fullSize && !isAbsolute && len(image.scales) != 0 {
		srcset = ScaleString(image.path, image.scales)
	}

	// determine link
	linkTarget := ""
	if image.link == "none" {
		// we're asked not to link to the image
		image.link = ""

	} else if image.link != "" {
		// link to something else

		// parse the link
		// ok, displaySame bool, target, display, tooltip, linkType string
		if ok, target, _, _, _ := page.parseLink(image.link, &fmtOpt{pos: image.getKeyPos("link")}); ok {
			image.link = target
			linkTarget = "_blank"
		} else {
			image.link = ""
		}

	} else {
		// link to the image

		image.link = image.path
	}

	// ############
	// ### HTML ###
	// ############

	// create an anchor for the link if there is one

	// this is just an image, no imagebox
	if !isBox {

		// put in link if there is one
		divOrA := el
		if image.link != "" {
			a := el.createChild("a", "image-a")
			a.setAttr("href", image.link)
			a.setAttr("target", linkTarget)
			divOrA = a
		}

		// create img with parent as either a or div
		img := divOrA.createChild("img", "image-img")
		img.setMeta("nonContainer", true)
		img.setAttr("src", image.path)
		img.setAttr("alt", image.alt)
		img.setAttr("srcset", srcset)

		return
	}

	// create inner box with width restriction
	inner := el.createChild("div", "imagebox-inner")
	inner.setStyle("width", strconv.Itoa(image.width)+"px")

	// put in link if there is one
	divOrA := inner
	if image.link != "" {
		a := inner.createChild("a", "image-a")
		a.setAttr("href", image.link)
		a.setAttr("target", linkTarget)
		divOrA = a
	}

	// create img with parent as either a or div
	img := divOrA.createChild("img", "imagebox-img")
	img.setMeta("nonContainer", true)
	img.setAttr("src", image.path)
	img.setAttr("alt", image.alt)
	img.setAttr("srcset", srcset)

	// insert javascript if using browser sizing
	if image.useJS {
		img.setAttr("onload", "quiki.imageResize(this);")
	}

	// description. we have to extract this here instead of in parse()
	// because at the time of parse() its text is not yet formatted
	desc, _ := image.Get("description")
	if desc == nil {
		desc, _ = image.Get("desc")
	}
	if desc != nil {
		inner.createChild(
			"div", "imagebox-description",
		).createChild(
			"div", "imagebox-description-inner",
		).add(desc)
	}
}

// fetch a string key, producing a warning at the appropriate spot if needed
func (image *imageBlock) getString(key string) string {
	s, err := image.GetStr(key)
	if err != nil {
		image.warn(image.getKeyPos(key), key+": "+err.Error())
		return ""
	}
	return s
}

// fetch a pixel size key, producing a warning at the appropriate spot if needed
func (image *imageBlock) getPx(key string) int {
	s, err := image.GetStr(key)
	if err != nil {
		image.warn(image.getKeyPos(key), key+": "+err.Error())
		return 0
	}
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(strings.TrimSuffix(s, "px"))
	if err != nil {
		image.warn(image.getKeyPos(key), key+": "+err.Error())
		return 0
	}
	return i
}
