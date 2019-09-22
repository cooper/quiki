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

			// trim again
			item = strings.Trim(item, "\t ")
			if item == "" {
				continue
			}

			// format, then trim again
			formatted := parseFormattedText(item)
			item = strings.Trim(string(formatted), "\t ")
			if item == "" {
				continue
			}

			el.addHtml(html(item))

		default:
			panic("not sure how to handle this content")
		}
	}
}
