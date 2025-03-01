package wikifier

type htmlBlock struct {
	*parserBlock
}

func newHTMLBlock(name string, b *parserBlock) block {
	return &htmlBlock{parserBlock: b}
}

func (b *htmlBlock) html(page *Page, el element) {
	el.setMeta("noIndent", true)
	el.setMeta("noTags", true)
	for _, item := range b.posContent() {
		content := item.content
		if s, ok := content.(string); ok {
			el.add(HTML(s))
		} else {
			el.add(content)
		}
	}
}
