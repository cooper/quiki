package wikifier

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var keyNormalizer = regexp.MustCompile(`\W`)
var keySplitter = regexp.MustCompile(`(.+)_(\d+)`)

// Map represents a Key-value dictionary.
// It is a quiki data type as well as the base of many block types.
type Map struct {
	didParse bool
	mapList  []*mapListEntry
	*parserBlock
	*variableScope
}

type mapListEntry struct {
	keyTitle string          // displayed key text
	key      string          // actual underlying key
	value    interface{}     // string, block, or mixed []interface{}
	typ      valueType       // value type
	pos      position        // position where the item started
	metas    map[string]bool // metadata
}

func (entry *mapListEntry) setMeta(key string, val bool) {
	if val == false {
		delete(entry.metas, key)
		return
	}
	entry.metas[key] = val
}

func (entry *mapListEntry) meta(key string) bool {
	return entry.metas[key]
}

type mapParser struct {
	key    interface{}
	values []interface{}

	escape        bool
	inValue       bool
	startPos      position
	pos           position
	overwroteKey  interface{}
	overwroteWith interface{}
	appendedKey   interface{}
}

// NewMap creates a new map, given the main block of the page it is to be associated with.
func NewMap(mb block) *Map {
	underlying := &parserBlock{
		openPos:      position{0, 0}, // FIXME
		parentB:      mb,
		parentC:      mb,
		typ:          "map",
		element:      newElement("div", "map"),
		genericCatch: &genericCatch{},
	}
	return &Map{false, nil, underlying, newVariableScope()}
}

func newMapBlock(name string, b *parserBlock) block {
	return &Map{false, nil, b, newVariableScope()}
}

func (m *Map) parse(page *Page) {

	// already parsed
	if m.didParse {
		return
	}
	m.didParse = true

	p := new(mapParser)
	for _, pc := range m.visiblePosContent() {
		p.pos = pc.position

		// infer start position to this one
		if p.startPos.none() {
			p.startPos = pc.position
		}

		switch item := pc.content.(type) {

		// block
		case block:

			if p.inValue {

				// first item
				if len(p.values) == 0 {
					p.startPos = p.pos
				}

				// add item
				p.values = append(p.values, item)

			} else {
				// overwrote a key
				p.overwroteKey = p.key
				p.overwroteWith = item
				p.key = item
			}
			m.warnMaybe(p)

			// parse the block
			item.parse(page)

		case string:
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			for i, c := range item {
				m.handleChar(i, p, c)
			}
		}
	}

	// positional warnings
	m.warnMaybe(p)
	keyHR, valueHR := humanReadableValue(p.key), humanReadableValue(p.values)

	// end of map warnings
	if valueHR != "" || p.inValue {
		// looks like we were in the middle of a value
		m.warn(p.pos, "Value "+valueHR+" for key "+keyHR+" not terminated")
	} else if keyHR != "" {
		// we were in the middle of a key
		m.warn(p.pos, "Stray key "+keyHR+" ignored")
	}

}

func (m *Map) handleChar(i int, p *mapParser, c rune) {

	if c == ':' && !p.inValue && !p.escape {
		// first colon indicates we're entering a value

		p.inValue = true

	} else if c == '\\' && !p.escape {
		// escape

		p.escape = true

	} else if c == ';' && !p.escape {
		// semicolon indicates termination of a pair

		strKey, isStrKey := p.key.(string)
		keyTitle := ""

		// determine key
		if (isStrKey && strKey == "") || p.key == nil {
			// this is something like
			//		: value; (can be text or block though)

			strKey = "anon_" + strconv.Itoa(i)
			p.key = strKey
			// no keyTitle

		} else if !p.inValue {
			// if there is a key but we aren't in the value,
			// it is something like
			//		value; (can be text or block though)

			// better to prefix text with : for less ambiguity
			if isStrKey && strKey[0] != '-' {
				m.warn(p.pos, "Standalone text should be prefixed with ':")
			}

			p.values = append(p.values, p.key)
			strKey = "anon_" + strconv.Itoa(i)
			p.key = strKey

			// no keyTitle

		} else {
			// otherwise it's a normal key-value pair
			// (can be text or block though)

			// we have to convert this to a string key somehow, so use the address
			if !isStrKey {
				strKey = fmt.Sprintf("%p", p.key)
			}

			// normalize the key for internal use
			strKey = strings.TrimSpace(strKey)
			keyTitle = strKey
			strKey = keyNormalizer.ReplaceAllString(strKey, "_")
			p.key = strKey

		}

		// fix the value
		// this returns either a string, block, or []interface{} of both
		// strings next to each other are merged; empty strings are removed
		valueToStore := fixValuesForStorage(p.values)

		// if this key exists, rename it to the next available <key>_key_<n>
		for exist, err := m.Get(strKey); exist != nil && err != nil; {
			matches := keySplitter.FindStringSubmatch(strKey)
			keyName, keyNumber := matches[1], matches[2]

			// first one, so make it _2
			if matches == nil {
				strKey += "_2"
				p.key = strKey
				continue
			}

			// it has _n, so increment that
			newKeyNumber, _ := strconv.Atoi(keyNumber)
			newKeyNumber++
			strKey = keyName + "_" + strconv.Itoa(newKeyNumber)
			p.key = strKey

		}

		// store the value in the underlying variableScope
		m.Set(strKey, valueToStore)

		// store the value in the map list
		m.mapList = append(m.mapList, &mapListEntry{
			keyTitle: keyTitle,                   // displayed key
			value:    valueToStore,               // string, block, or mixed []interface{}
			typ:      getValueType(valueToStore), // type of value
			key:      strKey,                     // actual underlying key
			pos:      p.startPos,                 // position where the item started
		})

		// check for warnings once more
		m.warnMaybe(p)

		// reset status
		p.inValue = false
		p.key = nil
		p.values = nil

	} else {
		// any other character; add to key or value
		p.escape = false

		// this is part of the value
		if p.inValue {

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

			return
		}

		// this is part of the key

		// starting a new key
		if p.key == nil {
			p.startPos = p.pos
			p.key = string(c)
			return
		}

		// check current key
		if lastStr, ok := p.key.(string); ok {
			// already working on a string key, so append it

			p.key = lastStr + string(c)
		} else if strings.TrimSpace(string(c)) != "" {
			// previous item was not a string
			// trying to add text to a non-text key...
			// (above ignores whitespace chars)

			p.appendedKey = p.key
		}
	}
}

// produce warnings as needed at current parser state
func (m *Map) warnMaybe(p *mapParser) {
	hrKey := humanReadableValue(p.key)

	// string keys spanning multiple lines are fishy
	if strKey, ok := p.key.(string); ok && strings.ContainsRune(strKey, '\n') {
		m.warn(p.pos, "Suspicious key "+hrKey)
	}

	// tried to append an object key
	if p.appendedKey != nil {
		appendText := humanReadableValue(p.appendedKey)
		m.warn(p.pos, "Stray text after "+appendText+" ignored")
		p.appendedKey = nil
	}

	// overwrote a key
	if p.overwroteKey != nil {
		old := humanReadableValue(p.overwroteKey)
		new := humanReadableValue(p.overwroteWith)
		m.warn(p.pos, "Overwrote "+old+" with "+new)
		p.overwroteKey = nil
		p.overwroteWith = nil
	}
}

func (m *Map) html(page *Page, el element) {
	for i, entry := range m.mapList {
		m.mapList[i].value = prepareForHTML(entry.value, page, entry.pos)
		m.setOwn(entry.key, m.mapList[i].value)
	}
}

func (m *Map) mainBlock() block {
	var b block = m
	for b.parentBlock() != nil {
		b = b.parentBlock()
	}
	return b
}
