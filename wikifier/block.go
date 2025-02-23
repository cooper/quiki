package wikifier

import (
	"fmt"
	"strings"
)

type block interface {
	String() string                    // description
	multi() bool                       // true when block produces multiple elements
	el() element                       // returns the html element
	parse(page *Page)                  // parse contents
	html(page *Page, el element)       // generate html element
	parentBlock() block                // parent block
	setParentBlock(p block)            // set parent block
	blockType() string                 // block type
	blockName() string                 // block name, if any
	close(pos Position)                // closes the block at the given position
	closed() bool                      // true when closed
	hierarchy() string                 // human-readable hierarchy
	blockContent() []block             // block children
	textContent() []string             // text children
	variables() *variableScope         // block variables or the page by default
	openPosition() Position            // position opened at
	warn(pos Position, warning string) // produce parser warning
	catch                              // all blocks must conform to catch
}

// generic base for all blocks
type parserBlock struct {
	typ, name string
	classes   []string
	openPos   Position
	closePos  Position
	parentB   block
	parentC   catch
	element   element
	headingID string
	_page     *Page // used for warnings
	*genericCatch
}

func (b *parserBlock) parse(page *Page) {
	// TODO: maybe split text nodes by line?
	// Note: this is not necessarily called for every block type inheriting Map.
	// Some blocks may parse() their children directly or not at all.
	for _, child := range b.blockContent() {
		child.parse(page)
	}
}

func (b *parserBlock) multi() bool {
	return false
}

func (b *parserBlock) el() element {
	return b.element
}

func (b *parserBlock) openPosition() Position {
	return b.openPos
}

func (b *parserBlock) parentBlock() block {
	return b.parentB
}

func (b *parserBlock) setParentBlock(p block) {
	b.parentB = p
}

func (b *parserBlock) blockType() string {
	return b.typ
}

func (b *parserBlock) blockName() string {
	return b.name
}

// can be overridden by blocks that have their own variable scope
func (b *parserBlock) variables() *variableScope {
	return b._page.variableScope
}

func (b *parserBlock) close(pos Position) {
	b.closePos = pos
}

func (b *parserBlock) closed() bool {
	return b.closePos.Line != 0 || b.closePos.Column != 0
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

func (b *parserBlock) catchType() catchType {
	return catchTypeBlock
}

func (b *parserBlock) runeOk(rune) bool {
	return true
}

func (b *parserBlock) shouldSkipRune(rune) bool {
	return false
}

func (b *parserBlock) warn(pos Position, warning string) {
	b._page.warn(pos, warning)
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

func (b *parserBlock) textContent() []string {
	var text []string
	for _, c := range b.content() {
		if txt, ok := c.(string); ok {
			text = append(text, txt)
		}
	}
	return text
}
