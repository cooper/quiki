package wikifier

// definitions{} is a definition list. each key is a term (<dt>) and its
// value the definition (<dd>).
type definitionsBlock struct {
	*Map
}

func newDefinitionsBlock(name string, b *parserBlock) block {
	return &definitionsBlock{newMapBlock("", b).(*Map)}
}

func (db *definitionsBlock) html(page *Page, el element) {
	db.Map.html(page, nil)
	el.setTag("dl")
	for _, entry := range db.mapList {
		if entry.keyTitle == "" {
			db.warn(entry.pos, "Definition without a term ignored")
			continue
		}
		dt := el.createChild("dt", "definitions-term")
		dt.addText(entry.keyTitle)
		dd := el.createChild("dd", "definitions-definition")
		dd.add(entry.value)
	}
}
