package wikifier

type historyBlock struct {
	*Map
}

func newHistoryBlock(name string, b *parserBlock) block {
	b.typ = "history"
	return &historyBlock{newMapBlock("", b).(*Map)}
}

func (b *historyBlock) html(page *Page, el element) {
	b.Map.html(page, el)
	table := el.createChild("table", "history-table")

	// append each pair
	for _, pair := range b.mapList {
		tr := table.createChild("tr", "history-pair")

		// key
		tr.createChild("td", "history-key").add(b.Fmt(pair.keyTitle, pair.pos))

		// value
		tr.createChild("td", "history-value").add(pair.value)
	}
}
