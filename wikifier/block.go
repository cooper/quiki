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

func (b *parserBlock) getParentCatch() parserCatch {
	return b.parent
}

func (b *parserBlock) catchType() string {
	return catchTypeBlock
}

func (b *parserBlock) byteOK(byte) bool {
	return true
}

func (b *parserBlock) shouldSkipByte(byte) bool {
	return false
}
