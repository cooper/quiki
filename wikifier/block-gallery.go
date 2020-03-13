package wikifier

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type galleryBlock struct {
	thumbHeight int
	images      []*galleryEntry
	*Map
}

type galleryEntry struct {
	thumbPath string
	img       *imageBlock
}

func newGalleryBlock(name string, b *parserBlock) block {
	return &galleryBlock{
		thumbHeight: 220,
		Map:         newMapBlock("", b).(*Map),
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
			g.thumbHeight = height

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

	// determine largest support retina scale
	// this will be used as the multiplier
	multi := 1
	for _, scale := range page.Opt.Image.Retina {
		if scale > multi {
			multi = scale
		}
	}

	// generate the thumbnail
	img.height = g.thumbHeight * multi
	img.parse(page)

	// add the image
	g.images = append(g.images, entry)
}

func (g *galleryBlock) html(page *Page, el element) {
	//g.Map.html(page, nil) -- skip since we don't want to convert to HTML, right?

	// create gallery options
	options := `{
		"thumbHeight": "` + strconv.Itoa(g.thumbHeight) + `",
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
	el.setAttr("id", "q-"+el.id())

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
		a.setAttr("data-ngthumb", entry.thumbPath)
		a.setAttr("data-ngdesc", desc)
	}
}
