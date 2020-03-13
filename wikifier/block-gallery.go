package wikifier

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type galleryBlock struct {
	thumbnailHeight int
	images          []*galleryEntry
	*Map
}

type galleryEntry struct {
	thumbnailPath string
	img           *imageBlock
}

func newGalleryBlock(name string, b *parserBlock) block {
	return &galleryBlock{
		thumbnailHeight: 320,
		Map:             newMapBlock("", b).(*Map),
	}
}

func (g *galleryBlock) parse(page *Page) {
	g.Map.parse(page)

	// sort out the map
	for _, imgKey := range g.OrderedKeys() {
		switch imgKey {

		// thumbnail height
		case "thumb_height":
			thumbHeight, err := g.GetStr(imgKey)

			// not a string
			if err != nil {
				g.warn(g.openPos, errors.Wrap(err, imgKey).Error()) // FIXME: use key position
				break
			}

			// convert to int
			height, err := strconv.Atoi(thumbHeight)
			if err != nil {
				g.warn(g.openPos, "thumb_height: expected integer")
				break
			}

			// good
			g.thumbnailHeight = height

		default:

			// unknown key
			if !strings.HasPrefix(imgKey, "anon_") {
				g.warn(g.openPos, "Invalid key '"+imgKey+"'") // FIXME: use key position
				break
			}

			// anonymous image is OK
			blk, err := g.GetBlock(imgKey)

			// non-block
			if err != nil {
				g.warn(g.openPos, errors.Wrap(err, imgKey).Error()) // FIXME: use key position
				break
			}

			// it is indeed a block, but is it an image?
			img, ok := blk.(*imageBlock)
			if !ok {
				// block other than image
				g.warn(g.openPos, imgKey+": expected Block<image{}>") // FIXME: use key position
				break
			}

			// it is indeed an image!
			g.addImage(page, img)
		}
	}
}

func (g *galleryBlock) addImage(page *Page, img *imageBlock) {

	// get full-size path
	img.path = page.Opt.Root.Image + "/" + img.file
	entry := &galleryEntry{img.path, img}

	// generate the thumbnail
	if page.Opt.Image.SizeMethod == "server" {

		// Calc and Sizer must be present
		if page.Opt.Image.Sizer == nil || page.Opt.Image.Calc == nil {
			img.warn(img.openPos, "image.sizer and image.calc required with image.size_method 'server'")
			img.parseFailed = true
			return
		}

		// calculate missing dimension
		width, height := page.Opt.Image.Calc(
			img.file,
			0,
			g.thumbnailHeight,
			page,
		)

		// create thumbnail path
		entry.thumbnailPath = page.Opt.Image.Sizer(
			img.file,
			width,
			height,
			page,
		)
	}

	// add the image
	g.images = append(g.images, entry)
}

func (g *galleryBlock) html(page *Page, el element) {
	//g.Map.html(page, nil) -- skip since we don't want to convert to HTML, right?

	// create gallery options
	options := `{
		"thumbnailHeight": "` + strconv.Itoa(g.thumbnailHeight) + `",
		"thumbnailWidth": "auto",
		"thumbnailBorderVertical": 0,
		"thumbnailBorderHorizontal": 0,
		"colorScheme": {
			"thumbnail": {
				"borderColor": "rgba(0,0,0,0)"
			}
		},
		"thumbnailDisplayTransition": "flipUp",
		"thumbnailDisplayTransitionDuration": 500,
		"thumbnailLabel": {
			"displayDescription": true,
			"descriptionMultiLine": true
		},
		"thumbnailHoverEffect2": "descriptionSlideUp",
		"thumbnailAlignment": "center",
		"thumbnailGutterWidth": 10,
		"thumbnailGutterHeight": 10
	}`

	// set options
	el.setAttr("data-nanogallery2", options)
	el.setMeta("needID", true)

	// add images
	for _, entry := range g.images {

		// determine desc
		// consider: this could be extracted in image{} parse.
		// I didn't do it since image{} usually didn't have a desc.
		desc, _ := entry.img.GetStr("description")
		if desc == "" {
			desc, _ = entry.img.GetStr("desc")
		}

		// create gallery item
		a := el.createChild("a", "")
		a.setAttr("href", entry.img.path)
		a.setAttr("data-ngthumb", entry.thumbnailPath)
		a.setAttr("data-ngdesc", desc)
	}
}
