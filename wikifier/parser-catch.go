package wikifier

import "log"

type parserCatch interface {
	getParent() parserCatch
	getPositionedContent() []positionedContent
	getContent() []interface{}
	getPrefixContent() []interface{}
	getLastString() string
	setLastContent(item interface{})
	appendContent(items []interface{}, pos parserPosition)
	shouldSkipByte(b byte) bool
	pushContent(item interface{}, pos parserPosition)
	pushContents(pc []positionedContent)
	appendString(s string, pos parserPosition)
}

type genericCatch struct {
	positionedContent       []positionedContent
	positionedPrefixContent []positionedContent
}

type positionedContent struct {
	content  interface{}
	position parserPosition
}

func (c *genericCatch) setLastContent(content interface{}) {
	c.positionedContent[len(c.positionedContent)-1].content = content
}

func (c *genericCatch) lastContent() interface{} {
	if c.positionedContent == nil {
		return nil
	}
	return c.positionedContent[len(c.positionedContent)-1].content
}

func (c *genericCatch) getLastString() string {
	if c.positionedContent == nil {
		return ""
	}
	if content, ok := c.lastContent().(string); !ok {
		return ""
	} else {
		return content
	}
}

// append any combination of blocks and strings
func (c *genericCatch) appendContent(content []interface{}, pos parserPosition) {
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
func (c *genericCatch) appendString(s string, pos parserPosition) {
	log.Printf("appendString: %v", s)

	// the location is empty, so this is the first item
	if len(c.positionedContent) == 0 {
		c.pushContent(s, pos)
		return
	}

	// append an existing string
	switch v := c.lastContent().(type) {
	case string:
		c.positionedContent[len(c.positionedContent)-1].content = v + s
	default:
		c.pushContent(s, pos)
	}
}

func (c *genericCatch) pushContent(item interface{}, pos parserPosition) {
	log.Printf("pushContent: %v/%v", item, pos)
	c.positionedContent = append(c.positionedContent, positionedContent{item, pos})
}

func (c *genericCatch) pushContents(pc []positionedContent) {
	log.Printf("pushContents: %v", pc)
	c.positionedContent = append(c.positionedContent, pc...)
}

func (c *genericCatch) shouldSkipByte(b byte) bool {
	return false
}

func (c *genericCatch) getPositionedContent() []positionedContent {
	return c.positionedContent
}

func (c *genericCatch) getContent() []interface{} {
	content := make([]interface{}, len(c.positionedContent))
	for i, pc := range c.positionedContent {
		content[i] = pc.content
	}
	return content
}

func (c *genericCatch) getPrefixContent() []interface{} {
	content := make([]interface{}, len(c.positionedPrefixContent))
	for i, pc := range c.positionedPrefixContent {
		content[i] = pc.content
	}
	return content
}
