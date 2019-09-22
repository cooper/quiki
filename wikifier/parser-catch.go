package wikifier

import "log"

const (
	catchTypeVariableName  = "Variable name"
	catchTypeVariableValue = "Variable value"
	catchTypeBlock         = "Block"
)

type catch interface {
	parentCatch() catch
	posContent() []posContent
	positionedPrefixContent() []posContent
	content() []interface{}
	prefixContent() []interface{}
	lastString() string
	setLastContent(item interface{})
	appendContent(items []interface{}, pos position)
	pushContent(item interface{}, pos position)
	pushContents(pc []posContent)
	appendString(s string, pos position)
	byteOK(b byte) bool
	shouldSkipByte(b byte) bool
	catchType() string
}

type genericCatch struct {
	positioned       []posContent
	positionedPrefix []posContent
}

type posContent struct {
	content  interface{}
	position position
}

func (c *genericCatch) setLastContent(content interface{}) {
	c.positioned[len(c.positioned)-1].content = content
}

func (c *genericCatch) lastContent() interface{} {
	if c.positioned == nil {
		return nil
	}
	return c.positioned[len(c.positioned)-1].content
}

func (c *genericCatch) lastString() string {
	if c.positioned == nil {
		return ""
	}
	if content, ok := c.lastContent().(string); !ok {
		return ""
	} else {
		return content
	}
}

// append any combination of blocks and strings
func (c *genericCatch) appendContent(content []interface{}, pos position) {
	log.Printf("appendContent: %v", content)
	for _, item := range content {
		switch v := item.(type) {
		case string:
			c.appendString(v, pos)
		default:
			c.pushContent(v, pos)
		}
	}
}

// append an existing string if the last item is one
func (c *genericCatch) appendString(s string, pos position) {
	log.Printf("appendString: %v", s)

	// the location is empty, so this is the first item
	if len(c.positioned) == 0 {
		c.pushContent(s, pos)
		return
	}

	// append an existing string
	switch v := c.lastContent().(type) {
	case string:
		c.positioned[len(c.positioned)-1].content = v + s
	default:
		c.pushContent(s, pos)
	}
}

func (c *genericCatch) pushContent(item interface{}, pos position) {
	log.Printf("pushContent: %v/%v", item, pos)
	c.positioned = append(c.positioned, posContent{item, pos})
}

func (c *genericCatch) pushContents(pc []posContent) {
	log.Printf("pushContents: %v", pc)
	c.positioned = append(c.positioned, pc...)
}

func (c *genericCatch) posContent() []posContent {
	return c.positioned
}

func (c *genericCatch) positionedPrefixContent() []posContent {
	return c.positionedPrefix
}

func (c *genericCatch) content() []interface{} {
	content := make([]interface{}, len(c.positioned))
	for i, pc := range c.positioned {
		content[i] = pc.content
	}
	return content
}

func (c *genericCatch) prefixContent() []interface{} {
	content := make([]interface{}, len(c.positionedPrefix))
	for i, pc := range c.positionedPrefix {
		content[i] = pc.content
	}
	return content
}
