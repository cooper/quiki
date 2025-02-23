package wikifier

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	keyNormalizer = regexp.MustCompile(`\W`)
	keySplitter   = regexp.MustCompile(`(.+)_(\d+)`)
)

// Map represents a Key-value dictionary.
// It is a quiki data type as well as the base of many block types.
type Map struct {
	noFormatValues bool
	didParse       bool
	mapList        []*mapListEntry
	*parserBlock
	*variableScope
}

type mapListEntry struct {
	keyTitle string          // displayed key text
	key      string          // actual underlying key
	value    any             // string, html, block, or mixed []any
	typ      valueType       // value type
	pos      Position        // position where the item started
	metas    map[string]bool // metadata
}

// func (entry *mapListEntry) setMeta(key string, val bool) {
// 	if !val {
// 		delete(entry.metas, key)
// 		return
// 	}
// 	entry.metas[key] = val
// }

func (entry *mapListEntry) meta(key string) bool {
	return entry.metas[key]
}

type mapParser struct {
	key    any
	values []any

	escape        bool
	inValue       bool
	startPos      Position
	pos           Position
	overwroteKey  any
	overwroteWith any
	appendedKey   any
}

// NewMap creates a new map, given the main block of the page it is to be associated with.
func NewMap(mb block) *Map {
	underlying := &parserBlock{
		openPos:      Position{0, 0}, // FIXME
		parentB:      mb,
		parentC:      mb,
		typ:          "map",
		element:      newElement("div", "map"),
		genericCatch: &genericCatch{},
	}
	return &Map{false, false, nil, underlying, newVariableScope()}
}

func newMapBlock(name string, b *parserBlock) block {
	return &Map{false, false, nil, b, newVariableScope()}
}

func (m *Map) parse(page *Page) {

	// already parsed
	if m.didParse {
		return
	}
	m.didParse = true

	p := new(mapParser)
	for i, pc := range m.posContent() {
		p.pos = pc.pos

		// infer start position to this one
		if p.startPos.none() {
			p.startPos = pc.pos
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
			item = strings.Trim(item, "\t ") // remove non-newline whitespace
			// item = strings.Replace(item, "\n", " ", -1) // convert newlines
			if item == "" {
				continue
			}
			for _, c := range item {
				m.handleChar(page, i, p, c)
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

func (m *Map) handleChar(_ *Page, i int, p *mapParser, c rune) {
	p.pos.Column = i

	if c == ':' && !p.inValue && !p.escape {
		// first colon indicates we're entering a value
		m.warnMaybe(p)
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
				m.warn(p.pos, "Standalone text should be prefixed with ':'")
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
		// this returns either a string, block, HTML, or []any combination
		// strings next to each other are merged; empty strings are removed
		valueToStore := fixValuesForStorage(p.values, m, p.pos, !m.noFormatValues)

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
			value:    valueToStore,               // string, block, or mixed []any
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

		// if it was escaped but not a parser char, add the \
		add := string(c)
		if p.escape && c != ';' && c != ':' && c != '\\' {
			add = "\\" + add
		}
		p.escape = false

		// this is part of the value
		if p.inValue {

			// first item
			if len(p.values) == 0 {
				p.startPos = p.pos
				p.values = append(p.values, add)
				return
			}

			// check previous item
			last := p.values[len(p.values)-1]
			if lastStr, ok := last.(string); ok {
				// previous item was a string, so append it

				p.values[len(p.values)-1] = lastStr + add
			} else {
				// previous item was not a string,
				// so start a new string item

				p.values = append(p.values, add)
			}

			return
		}

		// this is part of the key

		// starting a new key
		if p.key == nil {
			if strings.TrimSpace(add) == "" {
				// ignore whitespace at the start of keys
				return
			}
			p.startPos = p.pos
			p.key = add
			return
		}

		// check current key
		if lastStr, ok := p.key.(string); ok {
			// already working on a string key, so append it

			p.key = lastStr + add
		} else if strings.TrimSpace(add) != "" {
			// previous item was not a string
			// trying to add text to a non-text key...
			// (above ignores whitespace chars)

			p.appendedKey = p.key
		}
	}
}

// getEntry fetches the MapListEntry for a key.
func (m *Map) getEntry(key string) *mapListEntry {
	for _, entry := range m.mapList {
		if entry.key == key {
			return entry
		}
	}
	return nil
}

// getKeyPos returns the position where a key started.
// If the key doesn't exist, it returns the position where the map started.
func (m *Map) getKeyPos(key string) Position {
	if entry := m.getEntry(key); entry != nil {
		return entry.pos
	}
	return m.openPos
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

// default behavior for maps is to run html() on all children
// and replace the block value in the map with the generated element
func (m *Map) html(page *Page, el element) {
	if m.noFormatValues {
		return
	}
	for i, entry := range m.mapList {
		value := prepareForHTML(entry.value, m, entry.pos)
		m.mapList[i].value = value
		m.mapList[i].typ = getValueType(value)
		m.setOwn(entry.key, value)
	}
}

// since maps can be stored in variables and are generated on the fly,
// we sometimes need the main block to associate them with
func (m *Map) mainBlock() block {
	var b block = m
	for b.parentBlock() != nil {
		b = b.parentBlock()
	}
	return b
}

// Map returns the actual underlying Go map.
func (m *Map) Map() map[string]any {
	return m.vars
}

// Keys returns a string of actual underlying map keys.
func (m *Map) Keys() []string {
	keys := make([]string, len(m.vars))
	i := 0
	for key := range m.vars {
		keys[i] = key
		i++
	}
	return keys
}

// OrderedKeys returns a string of map keys in the order
// provided in the source. Keys that were set internally
// (and not from quiki source code) are omitted.
func (m *Map) OrderedKeys() []string {
	keys := make([]string, len(m.mapList))
	i := 0
	for _, entry := range m.mapList {
		keys[i] = entry.key
		i++
	}
	return keys
}
