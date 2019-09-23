package wikifier

import "strings"

type pBlock struct {
	*parserBlock
}

// TODO: headings, etc.

func newPBlock(name string, b *parserBlock) block {
	return &pBlock{parserBlock: b}
}

func (p *pBlock) parse(page *Page) {
}

func (p *pBlock) html(page *Page, el element) {
	el.setTag("p")

	for _, pc := range p.visiblePosContent() {
		switch item := pc.content.(type) {
		case block:
			item.html(page, item.el())
			el.addChild(item.el())

		case string:

			// trim again
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			// format, then trim again
			formatted := page.parseFormattedText(item)
			item = strings.TrimSpace(string(formatted))
			if item == "" {
				continue
			}

			el.addHtml(Html(item))

		default:
			panic("not sure how to handle this content")
		}
	}
}
