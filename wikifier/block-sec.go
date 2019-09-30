package wikifier

import (
	"fmt"
	"strconv"
	"strings"
)

type secBlock struct {
	n           int
	isIntro     bool
	headerLevel int
	*parserBlock
}

func newSecBlock(name string, b *parserBlock) block {
	return &secBlock{parserBlock: b}
}

func (sec *secBlock) parse(page *Page) {
	enable := page.Opt.Page.EnableTitle
	fmt.Printf("parse(): title(%v) sectionN(%v) enable(%v)\n", sec.blockName(), page.sectionN, enable)

	// @page.enable.title causes the first header to be larger than the
	// rest. it also uses @page.title as the first header if no other text
	// is provided.
	sec.n = page.sectionN
	sec.isIntro = sec.n == 0 && enable
	page.sectionN++

	// find the level from the parent section
	level := 1
	var blk block = sec
	for blk != nil {
		if parentSec, ok := blk.parentBlock().(*secBlock); ok {
			level = parentSec.headerLevel + 1
			break
		}
		blk = blk.parentBlock()
	}

	// top-level headers start at h2 when @page.enable.title is true, since the
	// page title is the sole h1. otherwise, h1 is top-level.
	if enable && level == 1 {
		level++
	}

	// intro is always h1
	if sec.isIntro {
		level = 1
	}

	// max is h6
	if level > 6 {
		level = 6
	}

	sec.headerLevel = level

	// this must come last so the section order is correct
	sec.parserBlock.parse(page)
}

func (sec *secBlock) html(page *Page, el element) {
	fmt.Printf("n(%v) isIntro(%v) title(%v)\n", sec.n, sec.isIntro, sec.blockName())
	// HEADING

	// determine if this is the intro section
	typ := "section-title"
	level := sec.headerLevel
	if sec.isIntro {
		typ = "section-page-title"
	}

	// use the page title if no other title is provided and @page.enable.title
	title, fmtTitle := sec.blockName(), HTML("")
	if sec.isIntro && title == "" {
		title = page.Title()
		fmtTitle = page.FmtTitle()
	}

	// we have a title
	if title != "" {

		// format title if we still need to
		if fmtTitle == "" {
			fmtTitle = page.parseFormattedTextOpts(title, &formatterOptions{pos: sec.openPos})
		}

		// TODO: meta section heading ID

		// heading ID
		headingID := PageNameLink(title, false)

		// add -n as needed if this is already used
		n := page.headingIDs[headingID]
		page.headingIDs[headingID]++
		if n != 0 {
			headingID += "-" + strconv.Itoa(n)
		}

		// create the heading
		h := el.createChild("h"+strconv.Itoa(level), typ)
		h.setAttr("id", "qa-"+headingID)
		h.addHTML(fmtTitle)
	}

	// CONTENT

	// iterate over content
	var contentToAdd []posContent
	for _, pc := range sec.posContent() {
		switch item := pc.content.(type) {
		case block:

			// create a section with the text up to this point
			sec.createParagraph(page, el, contentToAdd)
			contentToAdd = nil

			// adopt this block as my own
			item.html(page, item.el())
			el.addChild(item.el())

		case string:

			// if this is an empty line, create a new paragraph
			item = strings.TrimSpace(item)
			if item == "" {
				sec.createParagraph(page, el, contentToAdd)
				contentToAdd = nil
				continue
			}

			// otherwise, add it to the buffer
			contentToAdd = append(contentToAdd, pc)

		default:
			panic("not sure how to handle this content")
		}
	}

	// add whatever's left
	sec.createParagraph(page, el, contentToAdd)
}

func (sec *secBlock) createParagraph(page *Page, el element, pcs []posContent) {

	// this can be passed nothing
	if len(pcs) == 0 {
		return
	}

	// create a paragraph at first text node position
	p := newBlock("p", "", nil, sec, sec, pcs[0].position)
	p.appendContent(pcs, pcs[0].position)

	// parse and generate
	p.parse(page)
	p.html(page, p.el())

	// adopt it as my own
	el.addChild(p.el())
}
