package wikifier

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var keyNormalizer = regexp.MustCompile(`\W`)

type Map struct {
	*parserBlock
	*variableScope
}

type mapParser struct {
	key interface{}

	escape        bool
	inValue       bool
	startPos      position
	pos           position
	overwroteKey  interface{}
	overwroteWith interface{}
	appendedKey   interface{}
}

func NewMap(mb block) *Map {
	underlying := &parserBlock{
		openPos:      position{0, 0}, // FIXME
		parent:       mb,
		typ:          "map",
		element:      newElement("div", "map"),
		genericCatch: &genericCatch{},
	}
	return &Map{underlying, newVariableScope()}
}

func newMapBlock(name string, b *parserBlock) block {
	return &Map{b, newVariableScope()}
}

func (m *Map) parse(page *Page) {
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
				// append_value $value, $item, $pos, $startpos;

			} else {
				// overwrote a key
				p.overwroteKey = p.key
				p.overwroteWith = item
				p.key = item
			}
			m.warnMaybe(p)

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
}

func (m *Map) handleChar(i int, p *mapParser, c rune) {
	strKey, isStrKey := p.key.(string)

	if c == ':' && !p.inValue && !p.escape {
		// first colon indicates we're entering a value

		p.inValue = true

	} else if c == '\\' && !p.escape {
		// escape

		p.escape = true

	} else if c == ';' && !p.escape {
		// semicolon indicates termination of a pair

		keyTitle := ""

		// determine ky
		if (isStrKey && strKey == "") || p.key == nil {
			// this is something like
			//		: value;

			p.key = "anon_" + strconv.Itoa(i)
		} else if !p.inValue {
			// if there i a key but we aren't in the value,
			// it is something like
			//		value;

			if isStrKey && strKey[0] != '-' {
				m.warn(p.pos, "Standalone text should be prefixed with ':")
			}
		} else {
			// otherwise it's a normal key-value pair

			// we have to convert this to a string key somehow, so use the address
			if !isStrKey {
				strKey = fmt.Sprintf("%p", p.key)
			}

			// in any case, normalize the key for internal use
			strKey = strings.TrimSpace(strKey)
			keyTitle = strKey
			strKey = keyNormalizer.ReplaceAllString(strKey, "_")
			p.key = strKey

		}

		// fix the value
	}

	//             # fix the value
	//             fix_value $value;
	//             my $is_block = blessed $value; # true if ONE block and no text

	//             # if this key exists, rename it to the next available <key>_key_<n>.
	//             KEY: while (exists $values{$key}) {
	//                 my ($key_name, $key_number) = reverse map scalar reverse,
	//                     split('_', reverse($key), 2);
	//                 if (!defined $key_number || $key_number =~ m/\D/) {
	//                     $key = "${key}_2";
	//                     next KEY;
	//                 }
	//                 $key_number++;
	//                 $key = "${key_name}_${key_number}";
	//             }

	//             # store the value.
	//             $values{$key} = $value;
	//             push @{ $block->{map_array} }, {
	//                 key_title   => $key_title,     # displayed key
	//                 value       => $value,         # value, text or block
	//                 key         => $key,           # actual hash key
	//                 is_block    => $is_block,      # true if value was a block
	//                 pos         => { %$startpos }  # position
	//             };

	//             # warn bad keys and values
	//             $warn_bad_maybe->();

	//             # reset status.
	//             $in_value = 0;
	//             $key = $value = '';
	//         }

	//         # any other character
	//         else {
	//             $escaped = 0;

	//             # this is part of the value
	//             if ($in_value) {
	//                 append_value $value, $char, $pos, $startpos;
	//             }

	//             # this must be part of the key
	//             else {
	//                 if (blessed $key) {
	//                     $ap_key = $key unless $char =~ m/\s/;
	//                 }
	//                 else { $key .= $char }
	//             }
	//         }

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

func (m *Map) html(page *Page, el *element) {

}

func (m *Map) MainBlock() block {
	var b block = m
	for b.parentBlock() != nil {
		b = b.parentBlock()
	}
	return b
}
