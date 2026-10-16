package wikifier

import "strings"

// CSS generates and returns the CSS code for the page's inline styles.
func (p *Page) CSS() string {
	generated := ""
	for _, style := range p.styles {
		generated += p.cssSelector(style) + " {\n"
		for rule, value := range style.rules {
			generated += "    " + rule + ": " + value + ";\n"
		}
		generated += "}\n"
	}
	for _, style := range p.staticStyles {
		generated += style + "\n"
	}
	return generated
}

// elementSelector returns the selector matching one specific element.
// elements with an ID emit the class "q-<type>-<id>" (see generateIndented).
func elementSelector(el element) string {
	return ".q-" + cssEscape(el.elementType()+"-"+el.id())
}

// mainSelector returns the selector for this page's main element,
// which has no ID (and therefore only the class "q-main") on a nameless page
func (p *Page) mainSelector() string {
	main := p.main.el()
	if main.id() == "" {
		return ".q-main"
	}
	return elementSelector(main)
}

// cssSelector returns the full selector list for a style entry
func (p *Page) cssSelector(style styleEntry) string {

	// the element the style{} block belongs to. everything is scoped under this
	// page's main element, unless the parent is itself a main element (this
	// page's or a model's), which is already unique.
	parent := elementSelector(style.parent)
	if style.parent.elementType() != "main" {
		parent = p.mainSelector() + " " + parent
	}

	var parts []string
	if style.applyToParent {
		parts = append(parts, parent)
	}
	for _, set := range style.applyTo {
		parts = append(parts, parent+" "+p.cssSetString(set))
	}
	return strings.Join(parts, ",\n")
}

func (p *Page) cssSetString(set []string) string {
	items := make([]string, len(set))
	for i, item := range set {
		items[i] = p.cssItemString([]rune(item))
	}
	return strings.Join(items, " ")
}

func (p *Page) cssItemString(chars []rune) string {
	var str string
	var inClass, inID, inElType bool
	for _, char := range chars {

		switch char {

		// starting a class
		case '.':
			inClass = true
			str += ".qc-"

		// starting an ID
		case '#':
			inID = true
			str += ".qi-"

		// in neither a class nor an element type
		// assume that this is the start of element type
		default:
			if !inClass && !inID && !inElType && char != '*' {
				inElType = true
				str += ".q-"
			}
			str += string(char)

		}
	}

	// return $string;
	return str
}

// cssEscape escapes a class name for use in a selector. page names may contain
// characters like "/" and "." which are fine in a class attribute but would be
// misread in a selector.
func cssEscape(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r > 127:
			b.WriteRune(r)
		default:
			b.WriteByte('\\')
			b.WriteRune(r)
		}
	}
	return b.String()
}
