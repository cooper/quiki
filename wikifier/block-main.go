package wikifier

type mainBlock struct {
	*parserBlock
}

func newMainBlock(name string, b *parserBlock) block {
	return &mainBlock{parserBlock: b}
}

func (mb *mainBlock) parse(page *page) {

}

func (mb *mainBlock) html(page *page, el *element) {

}

func (mb *mainBlock) String() string {
	return "Main block"
}
