package wikifier

import (
	htmlfmt "html"
	"strconv"
	"strings"
)

var identifiers = make(map[string]int)

// HTML encapsulates a string to indicate that it is preformatted HTML.
// It lets quiki's parsers know not to attempt to format it any further.
type HTML string

type element interface {

	// ID
	id() string

	// tag/type
	tag() string
	setTag(tag string)
	elementType() string

	// attributes
	hasAttr(name string) bool
	attr(name string) string
	boolAttr(name string) bool
	setAttr(name, value string)
	setBoolAttr(name string, value bool)

	// styles
	hasStyle(name string) bool
	style(name string) string
	setStyle(name, value string)

	// metadata
	meta(name string) bool
	setMeta(name string, value bool)

	// adding content
	add(i interface{})
	addText(s string)
	addHTML(h HTML)
	addChild(child element)
	createChild(tag, typ string) element

	// classes
	addClass(class ...string)
	removeClass(class string) bool

	// parent
	parent() element
	setParent(parent element)

	// html generation
	generate() HTML
	generateIndented(indent int) []indentedLine
}

type genericElement struct {
	_tag          string                 // html tag
	_id           string                 // unique element identifier
	attrs         map[string]interface{} // html attributes
	styles        map[string]string      // inline styles
	metas         map[string]bool        // metadata
	typ           string                 // primary quiki class
	classes       []string               // quiki user-defined classes
	content       []interface{}          // mixed text and child elements
	parentElement element                // parent element, if any
	cachedHTML    HTML                   // cached version
	container     bool                   // true for container elements
}

func newElement(tag, typ string) element {
	identifiers[typ]++
	return &genericElement{
		_tag:      tag,
		_id:       typ + "-" + strconv.Itoa(identifiers[typ]),
		typ:       typ,
		container: true,
		attrs:     make(map[string]interface{}),
		styles:    make(map[string]string),
		metas:     make(map[string]bool),
	}
}

// fetch ID
func (el *genericElement) id() string {
	return el._id
}

// fetch tag
func (el *genericElement) tag() string {
	return el._tag
}

// set the tag
func (el *genericElement) setTag(tag string) {
	el._tag = tag
}

// fetch type
func (el *genericElement) elementType() string {
	return el.typ
}

// fetch string value for a meta
func (el *genericElement) meta(name string) bool {
	return el.metas[name]
}

// set string value for a meta
func (el *genericElement) setMeta(name string, value bool) {
	if value == false {
		delete(el.metas, name)
		return
	}
	el.metas[name] = value
}

// true when an attr is present on an element
func (el *genericElement) hasAttr(name string) bool {
	_, exist := el.attrs[name]
	return exist
}

// fetch string value for an attribute
func (el *genericElement) attr(name string) string {
	attr, exist := el.attrs[name]
	if !exist {
		return ""
	}
	if attrStr, ok := attr.(string); ok {
		return attrStr
	}
	return ""
}

// fetch boolean value for an attribute
func (el *genericElement) boolAttr(name string) bool {
	attr, exist := el.attrs[name]
	if !exist {
		return false
	}
	if attrBool, ok := attr.(bool); ok {
		return attrBool
	}
	return false
}

// set a string attribute
func (el *genericElement) setAttr(name, value string) {
	if value == "" {
		delete(el.attrs, name)
		return
	}
	el.attrs[name] = value
}

// set a boolean attribute
func (el *genericElement) setBoolAttr(name string, value bool) {
	if value == false {
		delete(el.attrs, name)
		return
	}
	el.attrs[name] = true
}

// true when a style key is present on an element
func (el *genericElement) hasStyle(name string) bool {
	_, exist := el.styles[name]
	return exist
}

// fetch string value for a style
func (el *genericElement) style(name string) string {
	return el.styles[name]
}

// set string value for a style
func (el *genericElement) setStyle(name, value string) {
	el.styles[name] = value
}

// add something
func (el *genericElement) add(i interface{}) {
	switch v := i.(type) {
	case string:
		el.addText(v)
	case HTML:
		el.addHTML(v)
	case element:
		el.addChild(v)
	case []interface{}:
		for _, val := range v {
			el.add(val)
		}
	case block:
		panic("add() block " + v.blockType() + "{} to element " + el.elementType())
	default:
		panic("add() unknown type to element " + el.elementType())
	}
}

// add a text node
func (el *genericElement) addText(s string) {
	el.content = append(el.content, s)
}

// add inner html
func (el *genericElement) addHTML(h HTML) {
	el.content = append(el.content, h)
}

// add child element
func (el *genericElement) addChild(child element) {
	el.content = append(el.content, child)
}

// create a child element and add it
func (el *genericElement) createChild(tag, typ string) element {
	child := newElement(tag, typ)
	el.addChild(child)
	return child
}

// fetch element's parent
func (el *genericElement) parent() element {
	return el.parentElement
}

// set this element's parent (internal only)
func (el *genericElement) setParent(parent element) {
	el.parentElement = parent // recursive!!
}

// add a class
func (el *genericElement) addClass(class ...string) {
	el.classes = append(el.classes, class...)
}

// remove a class, returning true if it was present
func (el *genericElement) removeClass(class string) bool {
	for i, v := range el.classes {
		if v == class {
			el.classes = append(el.classes[:i], el.classes[i+1:]...)
			return true
		}
	}
	return false
}

func (el *genericElement) generate() HTML {

	// cached version
	if el.cachedHTML != "" {
		return el.cachedHTML
	}

	el.cachedHTML = generateIndentedLines(el.generateIndented(0))
	return el.cachedHTML
}

type indentedLine struct {
	line   string
	indent int
}

func (el *genericElement) generateIndented(indent int) []indentedLine {
	var lines []indentedLine

	// if we haven't yet determined if this is a container,
	// check if it has any child elements
	if !el.container {
		el.container = len(el.content) != 0
	}

	// tags
	if !el.meta("noTags") {
		openingTag := "<" + el._tag

		// classes
		var classes []string
		if el.typ == "" {
			classes = make([]string, len(el.classes))
		} else {
			classes = make([]string, len(el.classes)+1)
			classes[0] = "q-" + el.typ
		}
		for i, name := range el.classes {
			if name[0] == '!' {
				name = name[1:]
				classes[i+1] = name
			} else {
				classes[i+1] = "q-" + name
			}
		}

		// inject ID
		if el.meta("needID") {
			classes = append([]string{"q-" + el._id}, classes...)
		}
		if len(classes) != 0 {
			openingTag += ` class="` + strings.Join(classes, " ") + `"`
		}

		// styles
		styles := ""
		for key, val := range el.styles {
			styles += key + ":" + val + "; "
		}
		if styles != "" {
			openingTag += ` style="` + styles + `"`
		}

		// other attributes
		for key, val := range el.attrs {
			switch v := val.(type) {
			case string:
				openingTag += " " + key + `="` + htmlfmt.EscapeString(v) + `"`
			case bool:
				openingTag += " " + key
			}
		}

		// non-container
		if !el.container {
			lines = append(lines, indentedLine{openingTag + " />", indent})
			return lines
		}

		// container
		lines = append(lines, indentedLine{openingTag + ">", indent})
	}

	// inner content
	for _, textOrEl := range el.content {
		var addLines []indentedLine
		switch v := textOrEl.(type) {

		case element:
			if v.meta("invisible") {
				break
			}
			if v.meta("noIndent") {
				addLines = v.generateIndented(0)
			} else {
				addLines = v.generateIndented(indent + 1)
			}

		case string:
			myIndent := indent + 1
			if el.meta("noIndent") {
				myIndent = 0
			}
			addLines = []indentedLine{indentedLine{htmlfmt.EscapeString(v), myIndent}}

		case HTML:
			myIndent := indent + 1
			if el.meta("noIndent") {
				myIndent = 0
			}
			addLines = []indentedLine{indentedLine{string(v), myIndent}}

		}

		lines = append(lines, addLines...)
	}

	// close it off
	if !el.meta("noTags") && !el.meta("noClose") {
		lines = append(lines, indentedLine{"</" + el._tag + ">", indent})
	}

	return lines
}

func generateIndentedLines(lines []indentedLine) HTML {
	generated := ""
	for _, line := range lines {
		generated += strings.Repeat("    ", line.indent) + line.line
		if line.line == "" || line.line[len(line.line)-1] != '\n' {
			generated += "\n"
		}
	}
	return HTML(generated)
}
