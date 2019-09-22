package wikifier

type mainBlock struct {
	*parserBlock
}

func newMainBlock(name string, b *parserBlock) block {
	return &mainBlock{parserBlock: b}
}

func (mb *mainBlock) parse(page *page) {

}

func (mb *mainBlock) html(page *page, el *element) {

	// always include the ID so that element styles can refer to it
	// (needed when more than 1 logical page is displayed in a browser window)
	el.needID = true

	// iterate over visible content only
	var contentToAdd []posContent
	for _, pc := range mb.visiblePosContent() {

		switch item := pc.content.(type) {
		case block:

			// create a section with the text up to this point
			mb.createSection(page, el, contentToAdd)
			contentToAdd = nil

			// adopt this block as my own
			item.html(page, item.el())
			el.addChild(item.el())

		case string:
			// TODO: trim the text and increment the line number appropriately
			if item == "" {
				continue
			}
			contentToAdd = append(contentToAdd, pc)

		default:
			panic("not sure how to handle this content")
		}
	}

	// add whatever's left
	mb.createSection(page, el, contentToAdd)
}

func (mb *mainBlock) createSection(page *page, el *element, pcs []posContent) {

	// this can be passed nothing
	if len(pcs) == 0 {
		return
	}

	// create a section at first text node position
	sec := newBlock("section", "", nil, mb, pcs[0].position)
	sec.pushContents(pcs)

	// parse and generate
	sec.parse(page)
	sec.html(page, sec.el())

	// adopt it as my own
	el.addChild(sec.el())
}
