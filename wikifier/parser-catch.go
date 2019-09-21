package wikifier

type parserCatch struct {
	content []interface{}
}

// append any combination of blocks and strings
func (c *parserCatch) appendContent(content []interface{}) {
	for _, item := range content {
		switch v := item.(type) {
		case string:
			c.appendStringIfString(v)
		default:
			c.pushContent(v)
		}
	}
}

// append an existing string if the last item is one
func (c *parserCatch) appendStringIfString(s string) {

	// the location is empty, so this is the first item
	if len(c.content) == 0 {
		c.pushContent(s)
		return
	}

	// append an existing string
	switch v := c.content[len(c.content)-1].(type) {
	case string:
		c.content[len(c.content)-1] = v + s
	default:
		c.pushContent(s)
	}
}

func (c *parserCatch) pushContent(item interface{}) {
	c.content = append(c.content, item)
}
