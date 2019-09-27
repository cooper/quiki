package wikifier

type modelBlock struct {
	*Map
}

func newModelBlock(name string, b *parserBlock) block {
	b.typ = "model"
	return &modelBlock{newMapBlock("", b).(*Map)}
}

func (mb *modelBlock) parse(page *Page) {
	mb.Map.parse(page)
}

func (mb *modelBlock) html(page *Page, el element) {
	mb.Map.html(page, el)

}
