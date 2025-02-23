package wikifier

type fmtBlock struct {
	*parserBlock
}

func newFmtBlock(name string, b *parserBlock) block {
	return &fmtBlock{parserBlock: b}
}

func (b *fmtBlock) html(page *Page, el element) {
	el.setMeta("noIndent", true)
	el.setMeta("noTags", true)
	for _, item := range b.posContent() {
		// if it's a string, format it
		if str, ok := item.content.(string); ok {
			el.add(formatOpts(b, str, item.pos, FmtOpt{NoEntities: true}))
			continue
		}
		el.add(item.content)
	}
}
