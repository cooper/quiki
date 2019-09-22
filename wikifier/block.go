package wikifier

import (
	"fmt"
	"strings"
)

type block interface {
	parentBlock() block
	blockType() string
	close(pos position)
	closed() bool
	catch
}

type parserBlock struct {
	typ, name string
	classes   []string
	openPos   position
	closePos  position
	parent    block
	*genericCatch
}

func (b *parserBlock) parentBlock() block {
	return b.parent
}

func (b *parserBlock) blockType() string {
	return b.typ
}

func (b *parserBlock) close(pos position) {
	b.closePos = pos
}

func (b *parserBlock) closed() bool {
	return b.closePos.line != 0 || b.closePos.column != 0
}

func (b *parserBlock) String() string {
	if b.name != "" {
		return fmt.Sprintf("Block<%s[%s]{}>", b.typ, b.name)
	}
	return fmt.Sprintf("Block<%s{}>", b.typ)
}

func (b *parserBlock) hierarchy() string {
	lines := []string{b.String()}
	for _, item := range b.getContent() {
		switch val := item.(type) {
		case string:
			lines = append(lines, val)
		case *parserBlock:
			split := strings.Split(val.hierarchy(), "\n")
			indented := make([]string, len(split))
			for i, blockLine := range split {
				indented[i] = "    " + blockLine
			}
			lines = append(lines, indented...)
		}
	}
	return strings.Join(lines, "\n")
}

func (b *parserBlock) getParentCatch() catch {
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
