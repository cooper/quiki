package wikifier

// infobox{}

// infobox{} displays a summary of information for an article.
// Usually there is just one per article, and it occurs before the first section.
type infobox struct {
	*Map
}

// newInfobox creates an infobox given an underlying parser block.
func newInfobox(name string, b *parserBlock) block {
	b.typ = "infobox"
	return &infobox{newMapBlock("", b).(*Map)}
}

// parse parses the infobox contents.
func (ib *infobox) parse(page *Page) {
	ib.Map.parse(page)
}

// html converts the contents of the infobox to HTML elements.
func (ib *infobox) html(page *Page, el element) {
	ib.Map.html(page, nil)
	el.setTag("table")

	// display the title if there is one
	if ib.name != "" {
		th := el.createChild("tr", "infobox-title").createChild("th", "")
		th.setAttr("colspan", "2")
		th.addHTML(page.Fmt(ib.name, ib.openPos))
	}

	// add the rows
	infoTableAddRows(ib, el, page, ib.mapList)
}

// infosec{}

// infosec{} allows you to organize an infobox{} into sections.
type infosec struct {
	*Map
}

// newInfosec creates an infosec{} given an underlying parser block.
func newInfosec(name string, b *parserBlock) block {
	b.typ = "infosec"
	return &infosec{newMapBlock("", b).(*Map)}
}

// infosec{} yields multiple elements (table rows)
func (is *infosec) multi() bool {
	return true
}

// parse parses the infosec contents.
func (is *infosec) parse(page *Page) {
	is.Map.parse(page)
}

// html converts the contents of the infosec to HTML elements.
func (is *infosec) html(page *Page, els element) {
	is.Map.html(page, nil)
	els.setMeta("isInfosec", true)

	// not in an infobox{}
	// FIXME: do not produce this warning if infosec{} is in a variable
	if is.parentBlock().blockType() != "infobox" {
		is.warn(is.openPosition(), "infosec{} outside of infobox{} does nothing")
		return
	}

	// inject the title
	if is.blockName() != "" {
		is.mapList = append([]*mapListEntry{&mapListEntry{
			key:   "_infosec_title_",
			metas: map[string]bool{"isTitle": true},
			value: page.Fmt(is.blockName(), is.openPosition()),
		}}, is.mapList...)
	}

	infoTableAddRows(is, els, page, is.mapList)
}

// INTERNALS

// infoTableAddRows appends each pair.
// Note that table might actually be an element collection.
func infoTableAddRows(infoboxOrSec block, table element, page *Page, pairs []*mapListEntry) {
	hasTitle := false

	// add a row for each entry
	for i, entry := range pairs {

		// if the value is from infosec{}, add each row
		if els, ok := entry.value.(element); ok && els.meta("isInfosec") {

			// infosec do not need a key
			if entry.keyTitle != "" {
				infoboxOrSec.warn(infoboxOrSec.openPosition(), "Key associated with infosec{} ignored")
			}

			table.addChild(els)
			continue
		}

		// determine next entry
		var next *mapListEntry
		hasNext := i != len(pairs)-1
		if hasNext {
			next = pairs[i+1]
		}

		var classes []string

		// this is the title
		isTitle := entry.meta("isTitle")
		if isTitle {
			classes = append(classes, "infosec-title")
			hasTitle = true
		}

		// this is the first item in this infosec
		if (hasTitle && i == 1) || (!hasTitle && i == 0) {
			classes = append(classes, "infosec-first")
		}

		// this is the last item in this infosec
		b4infosec := hasNext && next.meta("isInfobox")
		if !isTitle && (b4infosec || i == len(pairs)-1) {
			classes = append(classes, "infosec-last")
		}

		// not an infosec{}; this is a top-level pair
		infoTableAddRow(infoboxOrSec, table, entry, classes)
	}
}

// infoTableAddRow adds a row.
// Note that table might actually be an element collection.
func infoTableAddRow(infoboxOrSec block, table element, entry *mapListEntry, classes []string) {

	// create the row
	tr := table.createChild("tr", "infobox-pair")

	// append table row with a key
	if entry.keyTitle != "" {

		// key
		th := tr.createChild("th", "infobox-key")
		th.addText(entry.keyTitle)
		th.addClass(classes...)

		// value
		td := tr.createChild("td", "infobox-value")
		td.add(entry.value)
		td.addClass(classes...)

		return
	}

	// otherwise, append a table row without a key
	td := tr.createChild("td", "infobox-anon")
	td.setAttr("colspan", "2")
	td.add(entry.value)
	td.addClass(classes...)
}
