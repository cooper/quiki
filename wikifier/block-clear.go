package wikifier

type clearBlock struct {
	*parserBlock
}

func newClearBlock(name string, b *parserBlock) block {
	return &clearBlock{parserBlock: b}
}

func (b *clearBlock) parse(page *page) {

}

func (b *clearBlock) html(page *page, el *element) {

}
