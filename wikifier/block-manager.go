package wikifier

var blockAliases = map[string]string{
	"section":   "sec",
	"paragraph": "p",
	"hash":      "map",
	"format":    "fmt",
	"olist":     "numlist",
	"ulist":     "list",
}

var blockInitializers = map[string]func(name string, b *parserBlock) block{
	"main":      newMainBlock,
	"clear":     newClearBlock,
	"sec":       newSecBlock,
	"p":         newPBlock,
	"map":       newMapBlock,
	"infobox":   newInfobox,
	"infosec":   newInfosec,
	"invisible": newInvisibleBlock,
	"list":      newListBlock,
	"numlist":   newNumlistBlock,
	"code":      newCodeBlock,
	"fmt":       newFmtBlock,
	"html":      newHTMLBlock,
	"history":   newHistoryBlock,
	"style":     newStyleBlock,
	"imagebox":  newImagebox,
	"image":     newImageBlock,
	"model":     newModelBlock,
	"toc":       newTocBlock,
	"gallery":   newGalleryBlock,
	"for":       newForBlock,
}

func newBlock(blockType, blockName, headingID string, blockClasses []string, parentBlock block, parentCatch catch, pos Position, page *Page) block {
	if alias, exist := blockAliases[blockType]; exist {
		blockType = alias
	}
	el := newElement("div", blockType)
	for _, class := range blockClasses {
		el.addClass("!qc-" + class)
	}
	underlying := &parserBlock{
		openPos:      pos,
		parentB:      parentBlock,
		parentC:      parentCatch,
		typ:          blockType,
		name:         blockName,
		headingID:    headingID,
		classes:      blockClasses,
		element:      el,
		genericCatch: &genericCatch{},
		_variables:   page.variableScope,
		_page:        page,
	}
	if parentBlock != nil {
		underlying._variables = parentBlock.variables()
	}
	if init, ok := blockInitializers[blockType]; ok {
		b := init(blockName, underlying)

		// multi
		if b.multi() {
			underlying.element = newElements(nil)
		}

		return b
	}
	return newUnknownBlock(underlying)
}

func generateBlock(b block, page *Page) HTML {
	b.html(page, b.el()) // FIXME: actual page
	return b.el().generate()
}
