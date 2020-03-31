package wikifier

type pBlock struct {
	*parserBlock
}

func newPBlock(name string, b *parserBlock) block {
	return &pBlock{parserBlock: b}
}

func (p *pBlock) html(page *Page, el element) {
	el.setTag("p")

	for _, pc := range p.posContent() {
		switch item := pc.content.(type) {
		case block:
			item.html(page, item.el())
			el.addChild(item.el())

		case string:
			formatted := page.formatText(item, pc.pos)
			if item == "" {
				continue
			}
			el.addHTML(formatted)

		default:
			panic("not sure how to handle this content")
		}
	}
}
