package wikifier

import (
	"fmt"
	"log"
	"strings"
)

type catchType string

const (
	catchTypeVariableName  catchType = "Variable name"
	catchTypeVariableValue           = "Variable value"
	catchTypeBlock                   = "Block"
)

type catch interface {
	parentCatch() catch
	posContent() []posContent
	positionedPrefixContent() []posContent
	content() []interface{}
	prefixContent() []interface{}
	lastString() string
	setLastContent(item interface{})
	appendContent(item interface{}, pos position)
	appendContents(pc []posContent)
	byteOK(b byte) bool
	shouldSkipByte(b byte) bool
	catchType() catchType
}

type genericCatch struct {
	positioned       []posContent
	positionedPrefix []posContent

	line         string
	firstNewline bool
	removeIndent string
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
	content, ok := c.lastContent().(string)
	if !ok {
		return ""
	}
	return content
}

// append any combination of blocks and strings
func (c *genericCatch) appendContent(content interface{}, pos position) {
	switch v := content.(type) {
	case string:
		c.appendString(v, pos)
	case []posContent:
		c.appendContents(v)
	case []interface{}:
		for _, item := range v {
			c.appendContent(item, pos)
		}
	case posContent:
		c.appendContent(v.content, v.position)
	default:
		c.pushContent(v, pos)
	}
}

func (c *genericCatch) appendContents(pc []posContent) {
	log.Printf("pushContents: %v", pc)
	c.positioned = append(c.positioned, pc...)
}

// append an existing string if the last item is one
func (c *genericCatch) appendString(s string, pos position) {
	log.Printf("appendString: %v", s)
	c.line += s
	fmt.Println("line", c.line)

	if s[len(s)-1] == '\n' {
		if !c.firstNewline && len(c.line) > 2 {
			c.firstNewline = true // start a new one if the previous one ended in newline

			afterTrim := strings.TrimLeft(c.line, "\t ")
			difference := len(c.line) - len(afterTrim)
			if difference != 0 {
				c.removeIndent = c.line[:difference]
				// s = afterTrim
			}
			log.Printf("INDENT(%s) = (%s)", c.line, c.removeIndent)
		}
		fmt.Println("COMPLETE LINE:", strings.TrimPrefix(c.line, c.removeIndent))
		c.finishLine()
	}

	// the location is empty, so this is the first item
	if len(c.positioned) == 0 {
		c.pushContent(s, pos)
		return
	}

	// append an existing string
	switch v := c.lastContent().(type) {
	case string:
		if v != "" && v[len(v)-1] == '\n' {
			// start a new one if the previous one ended in newline
			c.pushContent(s, pos)
		} else {
			// other append the current string
			c.positioned[len(c.positioned)-1].content = v + s
		}
	default:
		c.pushContent(s, pos)
	}
}

func (c *genericCatch) finishLine() {
	c.line = ""

	// not working on a string..
	lastStr, ok := c.lastContent().(string)
	if !ok {
		return
	}

	// no indent magic to do, so that's it
	if c.removeIndent == "" {
		return
	}

	// scan backward to find where the line started
	// if there is no newline, it began at start of string
	lineStart := strings.LastIndexByte(lastStr, '\n')
	if lineStart == -1 {
		lineStart = 0
	}

	// trim the indent
	newPortion := strings.TrimPrefix(lastStr[lineStart:], c.removeIndent)
	newStr := lastStr[:lineStart] + newPortion
	c.positioned[len(c.positioned)-1].content = newStr
}

func (c *genericCatch) pushContent(item interface{}, pos position) {
	log.Printf("pushContent: %v/%v", item, pos)
	c.positioned = append(c.positioned, posContent{item, pos})
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
