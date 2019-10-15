package wikifier

type tocBlock struct {
	*parserBlock
}

func newTocBlock(name string, b *parserBlock) block {
	return &tocBlock{b}
}

func (toc *tocBlock) html(page *Page, el element) {
	el.setTag("ul")

	// add each top-level section
	for _, child := range page.main.blockContent() {
		if sec, ok := child.(*secBlock); ok {
			tocAdd(sec, el)
		}
	}
}

func tocAdd(sec *secBlock, addTo element) {

	// create an item for this section
	if !sec.isIntro {
		li := addTo.createChild("li", "")
		a := li.createChild("a", "link-internal")
		a.setAttr("href", "#"+sec.headingID)
		a.addText(sec.title)
		addTo = li
	}

	// create a sub-list for each section underneath
	var subList element
	for _, child := range sec.blockContent() {
		if secChild, ok := child.(*secBlock); ok {
			if subList == nil {
				subList = addTo.createChild("ul", "")
			}
			tocAdd(secChild, subList)
		}
	}
}
