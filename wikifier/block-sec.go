package wikifier

import (
	"strconv"
)

type secBlock struct {
	title       string
	fmtTitle    HTML
	n           int
	isIntro     bool
	headerLevel int
	*parserBlock
}

func newSecBlock(name string, b *parserBlock) block {
	return &secBlock{parserBlock: b}
}

func (sec *secBlock) parse(page *Page) {

	// wiki option
	enable := page.Opt.Page.EnableTitle

	// overwrite with local var if present
	val, _ := page.Get("page.enable.title")
	if boolVal, ok := val.(bool); ok {
		enable = boolVal
	}

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

	// determine section title
	// use the page title if no other title is provided and @page.enable.title
	sec.title, sec.fmtTitle = sec.blockName(), HTML("")
	if sec.isIntro && sec.title == "" {
		sec.title = page.Title()
		sec.fmtTitle = page.FmtTitle()
	}

	// determine heading ID
	// heading ID
	if sec.headingID == "" {
		sec.headingID = PageNameLink(sec.title)
	}

	// this must come last so the section order is correct
	sec.parserBlock.parse(page)
}

func (sec *secBlock) html(page *Page, el element) {
	// HEADING

	// determine if this is the intro section
	typ := "sec-title"
	level := sec.headerLevel
	if sec.isIntro {
		typ = "sec-page-title"
	}

	// we have a title
	if sec.title != "" {

		// format title if we still need to
		if sec.fmtTitle == "" {
			sec.fmtTitle = format(sec, sec.title, sec.openPos)
		}

		// TODO: meta section heading ID

		// add -n as needed if this is already used
		n := page.headingIDs[sec.headingID]
		page.headingIDs[sec.headingID]++
		if n != 0 {
			sec.headingID += "-" + strconv.Itoa(n)
		}

		// create the heading
		h := el.createChild("h"+strconv.Itoa(level), typ)
		h.setAttr("id", "qa-"+sec.headingID)
		h.addHTML(sec.fmtTitle)
	}

	handleGenericContent(sec, page, el)
}
