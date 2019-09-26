package wikifier

type htmlBlock struct {
	*parserBlock
}

func newHTMLBlock(name string, b *parserBlock) block {
	return &fmtBlock{parserBlock: b}
}

func (b *htmlBlock) html(page *Page, el element) {
	el.setMeta("noIndent", true)
	el.setMeta("noTags", true)
	for _, item := range b.posContent() {
		el.add(item.content)
	}
}
