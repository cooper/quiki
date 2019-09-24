package wikifier

type invisibleBlock struct {
	*parserBlock
}

func newInvisibleBlock(name string, b *parserBlock) block {
	return &invisibleBlock{b}
}

func (b *invisibleBlock) parse(page *Page)            {}
func (b *invisibleBlock) html(page *Page, el element) {}
