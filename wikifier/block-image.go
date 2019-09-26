package wikifier

import (
	"regexp"
	"strconv"
	"strings"
)

var externalImageRegex = regexp.MustCompile(`^(.+)://`)

type imageBlock struct {
	file, path, alt, link, lastName string
	align, float, author, license   string
	widthString, heightString       string
	width, height                   int
	parseFailed, useJS              bool
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
	var w, h int

	// fetch string values from map
	image.file = image.getString("file")
	image.alt = image.getString("alt")
	image.link = image.getString("link")
	image.align = image.getString("align")
	image.float = image.getString("float")
	image.author = image.getString("author")
	image.license = image.getString("license")

	// fetch dimensions
	image.width = image.getPx("width")
	image.height = image.getPx("height")

	// compatibility
	if image.align == "" {
		image.align = image.getString("float")
	}

	// no dimensions. if it's an inbox we can guess it
	if image.width == 0 && image.height == 0 && image.parentBlock().blockType() == "infobox" {
		image.width = 270
	}

	// no file - this is mandatory
	if image.file == "" {
		image.warn(image.openPos, "No file specified for image")
		image.parseFailed = true
		return
	}

	split := strings.Split(image.file, "/")
	image.path = image.file
	image.lastName = split[len(split)-1]

	// ##############
	// ### SIZING ###
	// ##############

	sizeMethod := strings.ToLower(page.Opt.Image.SizeMethod)

	if externalImageRegex.MatchString(image.file) {
		// if the file is an absolute URL, we cannot size the image
		// do nothing

	} else if image.width != 0 && image.height != 0 {
		// both dimensions were given, so we need to do no sizing.
		// FIXME: this forces the full size image instead of generating in the
		// given dimensions

		w = image.width
		h = image.height

	} else if sizeMethod == "javascript" {
		// use javascript image sizing
		//
		// uses full-size images directly and uses javascript to size imageboxes
		// - voids the validity as XHTML 1.0 Strict
		// - causes slight flash on page load when images are scaled

		// inject javascript resizer if no width is given
		if image.width == 0 {
			image.useJS = true
		}

		// width and height dummies will be overriden by JavaScript
		if image.width == 0 {
			w = 200
		} else {
			w = image.width
		}
		h = image.height

		// path is file relative to image root (full size image)
		image.path = page.Opt.Root.Image + "/" + image.file

	} else if sizeMethod == "server" {
		// use server-size image sizing
		//
		// - maintains XHTML 1.0 Strict validity
		// - eliminates flash on page load
		// - faster page load (since image files are smaller)
		// - require read access to local image directory

		// this must be provided by wiki
		if page.Opt.Image.Sizer == nil {
			image.warn(image.openPos, "image.sizer required with image.size_method 'server'")
			image.parseFailed = true
			return
		}

		// path is as returned by the function that sizes the image
		image.path = page.Opt.Image.Sizer(
			image.file,
			w,
			h,
			page,
		)

		// remember that we use this image in these dimensions on this page
		page.images[image.file] = append(page.images[image.file], []int{w, h})
	}

	// convert dimensions to string
	if w == 0 {
		image.widthString = "auto"
	} else {
		image.widthString = strconv.Itoa(w) + "px"
	}
	if h == 0 {
		image.heightString = "auto"
	} else {
		image.heightString = strconv.Itoa(h) + "px"
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

}

// FIXME: use position of the key
func (image *imageBlock) getString(key string) string {
	s, err := image.GetStr(key)
	if err != nil {
		image.warn(image.openPos, key+": "+err.Error())
		return ""
	}
	return s
}

// FIXME: use position of the key
func (image *imageBlock) getPx(key string) int {
	s, err := image.GetStr(key)
	if err != nil {
		image.warn(image.openPos, key+": "+err.Error())
		return 0
	}
	i, err := strconv.Atoi(strings.TrimSuffix(s, "px"))
	if err != nil {
		image.warn(image.openPos, key+": "+err.Error())
		return 0
	}
	return i
}
