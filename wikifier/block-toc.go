package wikifier

type tocBlock struct {
	secCount int
	*parserBlock
}

func newTocBlock(name string, b *parserBlock) block {
	return &tocBlock{0, b}
}

func (toc *tocBlock) html(page *Page, el element) {
	el.setTag("ul")
	el.addHTML(HTML("<li><strong>Contents</strong></li>"))

	// add each top-level section
	for _, child := range page.main.blockContent() {
		if sec, ok := child.(*secBlock); ok {
			toc.tocAdd(sec, el, page)
		}
	}

	// don't show the toc if there are <2 on the page
	if toc.secCount < 2 {
		el.hide()
	}
}

func (toc *tocBlock) tocAdd(sec *secBlock, addTo element, page *Page) {
	toc.secCount++

	// create an item for this section if it has a title and isn't intro
	var subList element
	if !sec.isIntro || sec.title == "" {
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
			toc.tocAdd(secChild, subList, page)
		}
	}
}
