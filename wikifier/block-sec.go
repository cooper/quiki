package wikifier

import "strings"

type secBlock struct {
	*parserBlock
}

// TODO: headings, etc.

func newSecBlock(name string, b *parserBlock) block {
	return &secBlock{parserBlock: b}
}

func (sec *secBlock) parse(page *Page) {
}

func (sec *secBlock) html(page *Page, el *element) {

	// iterate over visible content only
	var contentToAdd []posContent
	for _, pc := range sec.visiblePosContent() {
		switch item := pc.content.(type) {
		case block:

			// create a section with the text up to this point
			sec.createParagraph(page, el, contentToAdd)
			contentToAdd = nil

			// adopt this block as my own
			item.html(page, item.el())
			el.addChild(item.el())

		case string:
			// trim the text and increment the line number appropriately
			lines := 0
			item = strings.TrimLeftFunc(item, func(r rune) bool {
				if r == '\n' {
					lines++
					return true
				}
				return r == ' ' || r == '\t'
			})
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			pc.position.line += lines
			contentToAdd = append(contentToAdd, pc)

		default:
			panic("not sure how to handle this content")
		}
	}

	// add whatever's left
	sec.createParagraph(page, el, contentToAdd)
}

func (sec *secBlock) createParagraph(page *Page, el *element, pcs []posContent) {

	// this can be passed nothing
	if len(pcs) == 0 {
		return
	}

	// create a paragraph at first text node position
	p := newBlock("p", "", nil, sec, pcs[0].position)
	p.pushContents(pcs)

	// parse and generate
	p.parse(page)
	p.html(page, p.el())

	// adopt it as my own
	el.addChild(p.el())
}
