package wikifier

import (
	"strings"
)

type mainBlock struct {
	*parserBlock
}

func newMainBlock(name string, b *parserBlock) block {
	return &mainBlock{parserBlock: b}
}

func (mb *mainBlock) parse(page *Page) {

	// convert text to sections with paragraphs
	var newContent []posContent
	var contentToAdd []posContent
	var textStartPos position
	for _, pc := range mb.posContent() {
		switch item := pc.content.(type) {
		case block:

			// create a section with the text up to this point
			sec := mb.createSection(page, contentToAdd)
			contentToAdd = nil

			// adopt this block as my own
			if sec != nil {
				newContent = append(newContent, posContent{sec, textStartPos})
			}

			// now parse and add this block also
			item.parse(page)
			newContent = append(newContent, posContent{item, pc.position})

		case string:
			if strings.TrimSpace(item) == "" && len(contentToAdd) == 0 {
				continue
			}
			if contentToAdd == nil {
				textStartPos = pc.position
			}
			contentToAdd = append(contentToAdd, pc)

		default:
			panic("not sure how to handle this content")
		}
	}

	// add whatever's left
	sec := mb.createSection(page, contentToAdd)
	if sec != nil {
		newContent = append(newContent, posContent{sec, textStartPos})
	}

	// overwrite content
	mb.positioned = newContent
}

func (mb *mainBlock) html(page *Page, el element) {

	// always include the ID so that element styles can refer to it
	// (needed when more than 1 logical page is displayed in a browser window)
	el.setMeta("needID", true)

	// everything should be converted to blocks by now
	for _, item := range mb.blockContent() {
		item.html(page, item.el())
		el.addChild(item.el())
	}
}

func (mb *mainBlock) createSection(page *Page, pcs []posContent) block {

	// this can be passed nothing
	if len(pcs) == 0 {
		return nil
	}

	// create a section at first text node position
	sec := newBlock("sec", "", "", nil, mb, mb, pcs[0].position, page)
	sec.appendContent(pcs, pcs[0].position)

	// parse
	sec.parse(page)

	return sec
}
