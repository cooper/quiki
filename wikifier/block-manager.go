package wikifier

var blockAliases = map[string]string{
	"section":   "sec",
	"paragraph": "p",
	"hash":      "map",
	"format":    "fmt",
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
	"code":      newCodeBlock,
	"fmt":       newFmtBlock,
	"history":   newHistoryBlock,
	"style":     newStyleBlock,
	"imagebox":  newImagebox,
	"image":     newImageBlock,
	"model":     newModelBlock,
}

func newBlock(blockType, blockName string, blockClasses []string, parentBlock block, parentCatch catch, pos position) block {
	if alias, exist := blockAliases[blockType]; exist {
		blockType = alias
	}
	underlying := &parserBlock{
		openPos:      pos,
		parentB:      parentBlock,
		parentC:      parentCatch,
		typ:          blockType,
		name:         blockName,
		classes:      blockClasses, // TODO: add them to the element with prefix qc-
		element:      newElement("div", blockType),
		genericCatch: &genericCatch{},
	}
	if init, ok := blockInitializers[blockType]; ok {
		b := init(blockName, underlying)

		// multi
		if b.multi() {
			underlying.element = newElements(nil)
		}

		return b
	}
	return newUnknownBlock(blockName, underlying)
}

func generateBlock(b block, page *Page) HTML {
	b.html(page, b.el()) // FIXME: actual page
	return b.el().generate()
}
