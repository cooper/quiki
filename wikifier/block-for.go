package wikifier

import (
	"strings"
)

type forBlock struct {
	iterableName string
	itemName     string
	_scope       *variableScope
	*parserBlock
}

func newForBlock(name string, b *parserBlock) block {
	scope := newVariableScopeWithParent(b.parentBlock().variables())
	return &forBlock{"", "", scope, b}
}

func (b *forBlock) variables() *variableScope {
	return b._scope
}

func (b *forBlock) parse(p *Page) {

	// separate the iterable name from the item name
	expression := strings.TrimSpace(b.name)
	parts := strings.Split(expression, " as")
	b.iterableName = strings.TrimPrefix(strings.TrimSpace(parts[0]), "@")
	b.itemName = "value"
	if len(parts) > 1 {
		b.itemName = strings.TrimPrefix(strings.TrimSpace(parts[1]), "@")
	}

	// warn if missing iterable name
	if b.iterableName == "" {
		b.warn(b.openPos, "for{} block is missing iterable name; should be for [@var]")
		return
	}

	// warn if missing item name
	if b.itemName == "" {
		b.warn(b.openPos, "for{} block is missing item name after 'as'")
		return
	}

	// call parse() on children
	b.parserBlock.parse(p)
}

func (b *forBlock) html(p *Page, el element) {

	// do nothing, this was warned above
	if b.itemName == "" || b.iterableName == "" {
		return
	}

	iterable, err := b.variables().Get(b.iterableName)
	if err != nil {
		b.warn(b.openPos, "for{} block: "+err.Error())
		return
	}

	switch v := iterable.(type) {
	case *List:
		for i, item := range v.list {
			b.variables().setOwn(b.itemName, item)
			b.variables().setOwn("index", i)
			b.injectContent(el)
		}
	case *Map:
		for i, item := range v.mapList {
			b.variables().setOwn(b.itemName, item.value)
			b.variables().setOwn("key", item.key)
			b.variables().setOwn("index", i)
			b.injectContent(el)
		}
	default:
		b.warn(b.openPos, "for{} block: @"+b.iterableName+" is not iterable")
	}
}

func (b *forBlock) injectContent(el element) {
	for _, item := range b.posContent() {
		value := prepareForHTML(item.content, b, item.pos)
		el.add(value)
	}
}
