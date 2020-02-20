package wikifier

type tocBlock struct {
	*parserBlock
}

func newTocBlock(name string, b *parserBlock) block {
	return &tocBlock{b}
}

func (toc *tocBlock) html(page *Page, el element) {
	secCount := 0
	el.setTag("ul")
	el.addHTML(HTML("<li><strong>Contents</strong></li>"))

	// add each top-level section
	for _, child := range page.main.blockContent() {
		if sec, ok := child.(*secBlock); ok {
			tocAdd(sec, el, page)
			secCount++
		}
	}

	// don't show the toc if there are <2 on the page
	if secCount < 2 {
		el.hide()
	}
}

func tocAdd(sec *secBlock, addTo element, page *Page) {

	// create an item for this section
	var subList element
	if !sec.isIntro {
		li := addTo.createChild("li", "")
		a := li.createChild("a", "link-internal")
		a.setAttr("href", "#"+sec.headingID)
		a.addHTML(page.formatTextOpts(sec.title, fmtOpt{pos: sec.openPos}))
		addTo = li
	} else {
		subList = addTo
	}

	// create a sub-list for each section underneath
	for _, child := range sec.blockContent() {
		if secChild, ok := child.(*secBlock); ok {
			if subList == nil {
				subList = addTo.createChild("ul", "")
			}
			tocAdd(secChild, subList, page)
		}
	}
}
