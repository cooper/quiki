package wikifier

import "strings"

// handles mixed block context and formatted text
func handleGenericContent(sec block, page *Page, el element) {
	// iterate over content
	var contentToAdd []posContent
	for _, pc := range sec.posContent() {
		switch item := pc.content.(type) {
		case block:

			// create a section with the text up to this point
			createParagraph(sec, page, el, contentToAdd)
			contentToAdd = nil

			// adopt a copy of this element as my own
			// copied because it might be iterated over in for{} blocks
			copy := item.el().copy()
			item.html(page, copy)
			el.addChild(copy)

		case string:

			// if this is an empty line, create a new paragraph
			item = strings.TrimSpace(item)
			if item == "" {
				createParagraph(sec, page, el, contentToAdd)
				contentToAdd = nil
				continue
			}

			// otherwise, add it to the buffer
			contentToAdd = append(contentToAdd, pc)

		default:
			// FIXME: don't panic
			panic("not sure how to handle this content")
		}
	}

	// add whatever's left
	createParagraph(sec, page, el, contentToAdd)
}

// creates paragraphs for stray text nodes
func createParagraph(parent block, page *Page, el element, pcs []posContent) {

	// this can be passed nothing
	if len(pcs) == 0 {
		return
	}

	// create a paragraph at first text node position
	p := newBlock("p", "", "", nil, parent, parent, pcs[0].pos, page)
	p.appendContent(pcs, pcs[0].pos)

	// parse and generate
	p.parse(page)
	p.html(page, p.el())

	// adopt it as my own
	el.addChild(p.el())
}
