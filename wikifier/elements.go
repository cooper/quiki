package wikifier

// A collection of elements.
type elements struct {
	elements      []element
	metas         map[string]string
	cachedHTML    Html
	parentElement element
}

// Creates a collection of elements.
func newElements(els []element) *elements {
	return &elements{elements: els, metas: make(map[string]string)}
}

// If els is empty, returns an empty string.
// Otherwise, returns the first element's tag.
func (els *elements) tag() string {
	if len(els.elements) == 0 {
		return ""
	}
	return els.elements[0].tag()
}

// Sets the tag on all underlying elements.
func (els *elements) setTag(tag string) {
	for _, el := range els.elements {
		el.setTag(tag)
	}
}

// Reports whether a metadata key is present in this collection.
func (els *elements) hasMeta(name string) bool {
	_, exist := els.metas[name]
	return exist
}

// Fetches a value from the collection's metadata.
func (els *elements) meta(name string) string {
	return els.metas[name]
}

// Sets a value in the collection's metadata.
func (els *elements) setMeta(name, value string) {
	els.metas[name] = value
}

// Always returns false, as an element collection has no attributes.
func (els *elements) hasAttr(name string) bool {
	return false
}

// Panics. Cannot fetch attribute from an element collection.
func (els *elements) attr(name string) string {
	panic("unimplemented")
}

// Panics. Cannot fetch attribute from an element collection.
func (els *elements) boolAttr(name string) bool {
	panic("unimplemented")
}

// Sets a string attribute on all underlying elements.
func (els *elements) setAttr(name, value string) {
	for _, el := range els.elements {
		el.setAttr(name, value)
	}
}

// Sets a boolean attribute on all underlying elements.
func (els *elements) setBoolAttr(name string, value bool) {
	for _, el := range els.elements {
		el.setBoolAttr(name, value)
	}
}

// Panics. Cannot fetch styles from an element collection.
func (els *elements) hasStyle(name string) bool {
	panic("unimplemented")
}

// Panics. Cannot fetch styles from an element collection.
func (els *elements) style(name string) string {
	panic("unimplemented")
}

// Sets a style on all underlying elements.
func (els *elements) setStyle(name, value string) {
	for _, el := range els.elements {
		el.setStyle(name, value)
	}
}

// Panics. Cannot add text node to a collection of elements.
func (els *elements) addText(s string) {
	panic("unimplemented")
}

// Panics. Cannot add raw HTML to a collection of elements.
func (els *elements) addHtml(h Html) {
	panic("unimplemented")
}

// Adds another element.
func (els *elements) addChild(child element) {
	els.elements = append(els.elements, child)
}

// Creates an element and adds it.
func (els *elements) createChild(tag, typ string) element {
	child := newElement(tag, typ)
	els.addChild(child)
	return child
}

// Fetches the parent of this element collection.
func (els *elements) parent() element {
	return els.parentElement
}

// Sets the parent of this element collection.
func (els *elements) setParent(parent element) {
	els.parentElement = parent // recursive!!
}

// Sets whether to include underlying elements' unique IDs in generated HTML.
func (els *elements) setNeedID(need bool) {
	for _, el := range els.elements {
		el.setNeedID(need)
	}
}

// Adds a class to all underlying elements.
func (els *elements) addClass(class string) {
	for _, el := range els.elements {
		el.addClass(class)
	}
}

// Removes a class from all underlying elements.
func (els *elements) removeClass(class string) bool {
	oneTrue := false
	for _, el := range els.elements {
		if el.removeClass(class) {
			oneTrue = true
		}
	}
	return oneTrue
}

// Generates and returns HTML for the elements.
func (els *elements) generate() Html {
	generated := ""

	// cached version
	if els.cachedHTML != "" {
		return els.cachedHTML
	}

	// add each
	for _, el := range els.elements {
		generated += string(el.generate())
	}

	els.cachedHTML = Html(generated)
	return els.cachedHTML
}
