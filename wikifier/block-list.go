package wikifier

import (
	"strings"
)

// TODO: Make list comply to AttributedObject but only accept integer keys

// List represents a list of items.
// It is a quiki data type as well as the base of many block types.
type List struct {
	didParse bool
	list     []*listEntry
	*parserBlock
}

type listEntry struct {
	value interface{}       // string, block, or mixed []interface{}
	typ   valueType         // value type
	pos   position          // position where the item started
	metas map[string]string // metadata
}

func (entry *listEntry) setMeta(key, val string) {
	entry.metas[key] = val
}

func (entry *listEntry) meta(key string) string {
	return entry.metas[key]
}

type listParser struct {
	values   []interface{}
	escape   bool
	startPos position
	pos      position
}

// NewList creates a new list, given the main block of the page it is to be associated with.
func NewList(mb block) *List {
	underlying := &parserBlock{
		openPos:      position{0, 0}, // FIXME
		parentB:      mb,
		parentC:      mb,
		typ:          "list",
		element:      newElement("div", "list"),
		genericCatch: &genericCatch{},
	}
	return &List{false, nil, underlying}
}

func newListBlock(name string, b *parserBlock) block {
	return &List{false, nil, b}
}

func (l *List) parse(page *Page) {

	// already parsed
	if l.didParse {
		return
	}
	l.didParse = true

	p := new(listParser)
	for _, pc := range l.posContent() {
		p.pos = pc.position

		// infer start position to this one
		if p.startPos.none() {
			p.startPos = pc.position
		}

		switch item := pc.content.(type) {

		// block
		case block:

			// first item
			if len(p.values) == 0 {
				p.startPos = p.pos
			}

			// add item
			p.values = append(p.values, item)

			// parse the block
			item.parse(page)

		// string
		case string:

			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			for i, c := range item {
				l.handleChar(page, i, p, c)
			}
		}
	}

	// we were in the middle of an item
	if valueHR := humanReadableValue(p.values); valueHR != "" {
		// looks like we were in the middle of a value
		l.warn(p.pos, "Value "+valueHR+" not terminated")
	}
}

func (l *List) handleChar(page *Page, i int, p *listParser, c rune) {

	if c == '\\' && !p.escape {
		// escapes the next character
		p.escape = true

	} else if c == ';' && !p.escape {
		// terminates a value

		// store the value
		valueToStore := fixValuesForStorage(p.values, page, true)
		l.list = append(l.list, &listEntry{
			value: valueToStore,               // string, block, or mixed []interface{}
			typ:   getValueType(valueToStore), // type of value
			pos:   p.startPos,                 // position where the item started
		})

		// reset
		p.values = nil

	} else {
		// any other character

		p.escape = false

		// first item
		if len(p.values) == 0 {
			p.startPos = p.pos
			p.values = append(p.values, string(c))
			return
		}

		// check previous item
		last := p.values[len(p.values)-1]
		if lastStr, ok := last.(string); ok {
			// previous item was a string, so append it

			p.values[len(p.values)-1] = lastStr + string(c)
		} else {
			// previous item was not a string,
			// so start a new string item

			p.values = append(p.values, string(c))
		}
	}
}

func (l *List) html(page *Page, el element) {
	el.setTag("ul")
	for i, entry := range l.list {

		// prepare the value for inclusion in HTML element
		value := prepareForHTML(entry.value, page, entry.pos)
		l.list[i].value = value

		// create a list item
		el.createChild("li", "list-item").add(value)
	}
}

func (l *List) mainBlock() block {
	var b block = l
	for b.parentBlock() != nil {
		b = b.parentBlock()
	}
	return b
}
