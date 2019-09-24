package wikifier

type codeBlock struct {
	*parserBlock
}

func newCodeBlock(name string, b *parserBlock) block {
	return &codeBlock{parserBlock: b}
}

func (cb *codeBlock) html(page *Page, el element) {
	el.setTag("pre")
	el.setMeta("noIndent", true)

	// if there is a block name, it's the language for syntax highlighting
	if lang := cb.blockName(); lang != "" {
		el.addClass("!prettyprint")
		el.addClass("!lang-" + lang)
	}

	// add each text node
	for _, text := range cb.textContent() {
		el.addText(text)
	}
}
