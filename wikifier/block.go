package wikifier

import (
	"fmt"
	"log"
	"strings"
)

type block interface {
	String() string                    // description
	el() element                      // returns the html element
	parse(page *Page)                  // parse contents
	html(page *Page, el element)      // generate html element
	parentBlock() block                // parent block
	blockType() string                 // block type
	blockName() string                 // block name, if any
	close(pos position)                // closes the block at the given position
	closed() bool                      // true when closed
	hierarchy() string                 // human-readable hierarchy
	invisible() bool                   // true for invisible blocks (generate no html)
	visiblePosContent() []posContent   // visible text/blocks
	blockContent() []block             // block children
	warn(pos position, warning string) // produce parser warning
	catch                              // all blocks must conform to catch
}

// generic base for all blocks
type parserBlock struct {
	typ, name string
	classes   []string
	openPos   position
	closePos  position
	parentB   block
	parentC   catch
	element   element
	*genericCatch
}

func (b *parserBlock) parse(page *Page) {
	// TODO: maybe split text nodes by line?
	for _, child := range b.blockContent() {
		child.parse(page)
	}
}

func (b *parserBlock) el() element {
	return b.element
}

func (b *parserBlock) parentBlock() block {
	return b.parentB
}

func (b *parserBlock) blockType() string {
	return b.typ
}

func (b *parserBlock) blockName() string {
	return b.name
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
	for _, item := range b.content() {
		switch val := item.(type) {
		case string:
			lines = append(lines, val)
		case block:
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

func (b *parserBlock) parentCatch() catch {
	return b.parentC
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

func (b *parserBlock) invisible() bool {
	return false
}

func (b *parserBlock) warn(pos position, warning string) {
	log.Printf("WARNING: %s{} at %v: %s", b.blockType(), pos, warning)
}

func (b *parserBlock) blockContent() []block {
	var blocks []block
	for _, c := range b.content() {
		if blk, ok := c.(block); ok {
			blocks = append(blocks, blk)
		}
	}
	return blocks
}

func (b *parserBlock) visiblePosContent() []posContent {
	content := make([]posContent, 0)
	for _, pc := range b.posContent() {
		if blk, ok := pc.content.(block); ok && blk.invisible() {
			continue
		}
		content = append(content, pc)
	}
	return content
}
