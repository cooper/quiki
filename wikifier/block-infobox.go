package wikifier

type infobox struct {
	*Map
}

func newInfobox(name string, b *parserBlock) block {
	b.typ = "infobox"
	return &infobox{newMapBlock("", b).(*Map)}
}

func (ib *infobox) parse(page *Page) {
	ib.Map.parse(page)
}

func (ib *infobox) html(page *Page, el element) {
	ib.Map.html(page, nil)
	el.setTag("table")

	// display the title if there is one
	if ib.name != "" {
		th := el.createChild("tr", "infobox-title").createChild("th", "")
		th.setAttr("colspan", "2")
		th.addHtml(page.parseFormattedTextOpts(ib.name, &formatterOptions{pos: ib.openPos}))
	}

	// add the rows
	infoboxTableAddRows(ib, el, page, ib.mapList)
}

// Appends each pair.
// Note that table might actually be an element collection.
func infoboxTableAddRows(infoboxOrSec block, table element, page *Page, pairs []*mapListEntry) {
	hasTitle := false

	// add a row for each entry
	for i, entry := range pairs {

		// if the value is from infosec{}, add each row
		if els, ok := entry.value.(element); ok && entry.meta("infosec") != "" {

			// infosec do not need a key
			if entry.keyTitle != "" {
				infoboxOrSec.warn(infoboxOrSec.openPosition(), "Key associated with infosec{} ignored")
			}

			table.addChild(els)
		}

		// determine next entry
		var next *mapListEntry
		hasNext := i != len(pairs)-1
		if hasNext {
			next = pairs[i+1]
		}

		var classes []string

		// this is the title
		isTitle := entry.meta("isTitle") != ""
		if isTitle {
			classes = append(classes, "infosec-title")
			hasTitle = true
		}

		// this is the first item in this infosec
		if (hasTitle && i == 1) || (!hasTitle && i == 0) {
			classes = append(classes, "infosec-first")
		}

		// this is the last item in this infosec
		b4infosec := hasNext && next.meta("isInfobox") != ""
		if !isTitle && (b4infosec || i == len(pairs)-1) {
			classes = append(classes, "infosec-last")
		}

		// not an infosec{}; this is a top-level pair
		infoboxTableAddRow(infoboxOrSec, table, entry, classes)
	}
}

// Adds a row.
// Note that table might actually be an element collection.
func infoboxTableAddRow(infoboxOrSec block, table element, entry *mapListEntry, classes []string) {

	// create the row
	tr := table.createChild("tr", "infobox-pair")

	// append table row with a key
	if entry.keyTitle != "" {

		// key
		th := tr.createChild("th", "infobox-key")
		th.addText(entry.keyTitle)
		th.addClasses(classes)

		// value
		td := tr.createChild("td", "infobox-value")
		td.add(entry.value)
		td.addClasses(classes)

		return
	}

	// otherwise, append a table row without a key
	td := tr.createChild("td", "infobox-anon")
	td.setAttr("colspan", "2")
	td.add(entry.value)
	td.addClasses(classes)
}
