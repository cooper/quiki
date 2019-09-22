package wikifier

import (
	"fmt"
)

type parserBlock struct {
	parser    *parser
	typ, name string
	classes   []string
	openPos   parserPosition
	closePos  parserPosition
	closed    bool
	parent    *parserBlock
	*genericCatch
}

func (b *parserBlock) String() string {
	if b.name != "" {
		return fmt.Sprintf("Block<%s[%s]{}>", b.typ, b.name)
	}
	return fmt.Sprintf("Block<%s{}>", b.typ)
}

func (b *parserBlock) getParent() parserCatch {
	return b.parent
}
