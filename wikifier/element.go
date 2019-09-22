package wikifier

import (
	htmlfmt "html"
	"strconv"
	"strings"
)

var identifiers = make(map[string]int)

type html string

type element struct {
	tag        string                 // html tag
	attr       map[string]interface{} // html attributes
	style      map[string]string      // inline styles
	id         string                 // unique element identifier
	typ        string                 // primary quiki class
	classes    []string               // quiki user-defined classes
	content    []interface{}          // mixed text and child elements
	parent     *element               // parent element, if any
	cachedHTML html                   // cached version
	container  bool                   // true for container elements
	needID     bool                   // true if we should include id
	noTags     bool                   // if true, only generate inner HTML
	noIndent   bool                   // if true, do not indent contents (for <pre>)
	noClose    bool                   // if true, do not close (containers only)
}

func newElement(tag, typ string) *element {
	identifiers[typ]++
	return &element{
		tag:       tag,
		id:        typ + "-" + strconv.Itoa(identifiers[typ]),
		typ:       typ,
		container: tag == "div",
	}
}

func (el *element) setAttr(name, value string) {
	el.attr[name] = value
}

func (el *element) setBoolAttr(name string, value bool) {
	if value == false {
		delete(el.attr, name)
		return
	}
	el.attr[name] = true
}

func (el *element) addText(s string) {
	el.content = append(el.content, s)
}

func (el *element) addHtml(h html) {
	el.content = append(el.content, h)
}

func (el *element) addChild(child *element) {
	child.parent = el // recursive!!
	el.content = append(el.content, child)
}

func (el *element) createChild(tag, typ string) *element {
	child := newElement(tag, typ)
	el.addChild(child)
	child.parent = el // recursive!!
	return child
}

func (el *element) addClass(class string) {
	el.classes = append(el.classes, class)
}

func (el *element) removeClass(class string) bool {
	for i, v := range el.classes {
		if v == class {
			el.classes = append(el.classes[:i], el.classes[i+1:]...)
			return true
		}
	}
	return false
}

func (el *element) generate() html {
	generated := ""

	// cached version
	if el.cachedHTML != "" {
		return el.cachedHTML
	}

	// if we haven't yet determined if this is a container,
	// check if it has any child elements
	if !el.container {
		el.container = len(el.content) != 0
	}

	// tags
	if !el.noTags {
		generated = "<" + el.tag

		// classes
		classes := make([]string, len(el.classes)+1)
		classes[0] = "q-" + el.typ
		for i, name := range el.classes {
			classes[i+1] = "qc-" + name
		}

		// inject ID
		if el.needID {
			classes = append([]string{"q-" + el.id}, classes...)
		}
		generated += ` class="` + strings.Join(classes, " ") + `"`

		// styles
		styles := ""
		for key, val := range el.style {
			styles += key + ":" + val + "; "
		}
		if styles != "" {
			generated += ` style="` + styles + `"`
		}

		// other attributes
		for key, val := range el.attr {
			switch v := val.(type) {
			case string:
				generated += " " + key + `="` + htmlfmt.EscapeString(v) + `"`
			case bool:
				generated += " " + key
			}
		}
	}

	// non-container
	if !el.container {
		generated += " />\n"
		el.cachedHTML = html(generated)
		return html(generated)
	}

	// inner content
	generated += ">\n"
	for _, textOrEl := range el.content {
		add := ""
		switch v := textOrEl.(type) {
		case html:
			add = string(v)
		case string:
			add = htmlfmt.EscapeString(v)
		case *element:
			add = string(v.generate())
		}
		if !el.noIndent {
			add = indent(add)
		}
		generated += add
	}

	// close it off
	if !el.noTags && !el.noClose {
		generated += "</" + el.tag + ">\n"
	}

	el.cachedHTML = html(generated)
	return html(generated)
}

func indent(str string) string {
	var res []rune
	bol := true
	for _, c := range str {
		if bol && c != '\n' {
			res = append(res, []rune("    ")...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}
