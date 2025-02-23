package wikifier

import (
	"strings"
)

type forBlock struct {
	iterableName string
	itemName     string
	invalid      bool
	_scope       *variableScope
	*parserBlock
}

func newForBlock(name string, b *parserBlock) block {
	scope := newVariableScopeWithParent(b.parentBlock().variables())
	return &forBlock{"", "", false, scope, b}
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
		b.warn(b.openPos, "Missing iterable name, should be: for [@var] {...}")
		b.invalid = true
		return
	}

	// warn if missing item name
	if b.itemName == "" {
		b.warn(b.openPos, "Missing item name after 'as', should be: for [@var as @item] {...}")
		b.invalid = true
		return
	}

	found, err := b.variables().Get(b.iterableName)
	if err != nil {
		b.warn(b.openPos, err.Error())
		b.invalid = true
		return
	}
	if found == nil {
		b.warn(b.openPos, "@"+b.iterableName+" is not defined")
		b.invalid = true
		return
	}

	// call parse() on children
	b.parserBlock.parse(p)
}

func (b *forBlock) html(p *Page, el element) {
	el.setMeta("noTags", true)
	el.setMeta("noIndent", true)

	// do nothing if invalid
	if b.invalid {
		return
	}

	iterable, err := b.variables().Get(b.iterableName)
	if err != nil {
		b.warn(b.openPos, err.Error())
		return
	}

	switch v := iterable.(type) {
	case *List:
		for i, item := range v.list {
			b.variables().setOwn(b.itemName, item.value)
			b.variables().setOwn("index", i)
			handleGenericContent(b, p, el)
		}
	case *Map:
		for i, item := range v.mapList {
			b.variables().setOwn(b.itemName, item.value)
			b.variables().setOwn("key", item.keyTitle)
			b.variables().setOwn("index", i)
			handleGenericContent(b, p, el)
		}
	default:
		b.warn(b.openPos, "@"+b.iterableName+" is not iterable")
	}
}
