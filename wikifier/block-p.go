package wikifier

import "strings"

type pBlock struct {
	*parserBlock
}

// TODO: headings, etc.

func newPBlock(name string, b *parserBlock) block {
	return &pBlock{parserBlock: b}
}

func (p *pBlock) parse(page *page) {
}

func (p *pBlock) html(page *page, el *element) {
	el.tag = "p"

	for _, pc := range p.visiblePosContent() {
		switch item := pc.content.(type) {
		case block:
			item.html(page, item.el())
			el.addChild(item.el())

		case string:
			item = strings.Trim(item, "\t ")
			if item == "" {
				continue
			}
			// TODO: format, then trim formatted text
			// then .addHtml()
			el.addText(item)

		default:
			panic("not sure how to handle this content")
		}
	}
}
