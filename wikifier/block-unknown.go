package wikifier

type unknownBlock struct {
	*parserBlock
}

func newUnknownBlock(name string, b *parserBlock) block {
	return &unknownBlock{parserBlock: b}
}

func (b *unknownBlock) parse(page *page) {

}

func (b *unknownBlock) html(page *page, el *element) {

}
