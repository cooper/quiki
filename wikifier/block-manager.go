package wikifier

var blockInitializers = map[string]func(name string, b *parserBlock) block{
	"main":  newMainBlock,
	"clear": newClearBlock,
}

func newBlock(blockType, blockName string, blockClasses []string, parent block, pos position) block {
	underlying := &parserBlock{
		openPos:      pos,
		parent:       parent,
		typ:          blockType,
		name:         blockName,
		classes:      blockClasses,
		element:      newElement("div", blockType),
		genericCatch: &genericCatch{},
	}
	if init, ok := blockInitializers[blockType]; ok {
		return init(blockName, underlying)
	}
	return newUnknownBlock(blockName, underlying)
}
