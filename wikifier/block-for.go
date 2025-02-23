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
	expression := strings.TrimSpace(b.name)
	parts := strings.Split(expression, " as ")
	b.iterableName = strings.TrimPrefix(parts[0], "@")
	b.itemName = "value"
	if len(parts) > 1 {
		b.itemName = strings.TrimPrefix(parts[1], "@")
	}
	if b.itemName == "" {
		b.warn(b.openPos, "for{} block is missing item name after 'as'")
		return
	}
	if b.iterableName == "" {
		b.warn(b.openPos, "for{} block is missing iterable name; should be for [@var]")
		return
	}
	b.parserBlock.parse(p)
}

func (b *forBlock) html(p *Page, el element) {
	// if b.itemName == "" || b.iterableName == "" {
	// 	return
	// }
	// iterable, ok := p.variables[varName]
	// if !ok {
	// 	return
	// }

	// switch v := iterable.(type) {
	// case *List:
	// 	for i, item := range v.list {
	// 		p.variables[varName] = item
	// 		if otherName != "" {
	// 			p.variables[otherName] = i
	// 		}
	// 		b.parserBlock.html(p, el)
	// 	}
	// case *Map:
	// 	for key, value := range v.Items {
	// 		p.variables[varName] = value
	// 		if otherName != "" {
	// 			p.variables[otherName] = key
	// 		}
	// 		b.parserBlock.html(p, el)
	// 	}
	// }
}
