package wikifier

import "strings"

// CSS generates and returns the CSS code for the page's inline styles.
func (p *Page) CSS() string {
	generated := ""
	for _, style := range p.styles {
		applyTo := p.cssApplyString(style.applyTo)
		generated += applyTo + " {\n"
		for rule, value := range style.rules {
			generated += "    " + rule + ": " + value + ";\n"
		}
		generated += "}\n"
	}
	return generated
}

func (p *Page) cssApplyString(sets [][]string) string {
	parts := make([]string, len(sets))
	for i, set := range sets {
		str := p.cssSetString(set)
		if !strings.HasPrefix(str, ".q-main") {
			id := p.main.el().id()
			str = ".q-" + id + " " + str
		}
		parts[i] = str
	}
	return strings.Join(parts, ",\n")
}

func (p *Page) cssSetString(set []string) string {
	for i, item := range set {
		set[i] = p.cssItemString([]rune(item))
	}
	return strings.Join(set, " ")
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
