package wikifier

type unknownBlock struct {
	*parserBlock
}

func newUnknownBlock(b *parserBlock) block {
	return &unknownBlock{parserBlock: b}
}

func (b *unknownBlock) parse(page *Page) {

}

func (b *unknownBlock) html(page *Page, el element) {

}
