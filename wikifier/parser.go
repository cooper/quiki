package wikifier

import (
	"errors"
	"regexp"
	"strings"
)

type parser struct {
	line   int
	column int

	this    byte // current byte
	next    byte // next byte
	skip    bool // skip next byte
	escaped bool

	catch *parserCatch
	block *parserBlock

	ignoreLevel  int
	commentLevel int
	braceLevel   int
	braceFirst   bool

	lastContent []byte
}

func (p parser) parseLine(line []byte) error {
	for i, b := range line {

		// skip this byte
		if p.skip {
			p.skip = false
			continue
		}

		// update column and characters
		p.column = i
		p.this = b
		p.next = line[i+1]

		// handle this character and give up if error occurred
		if err := p.parseByte(b); err != nil {
			return err
		}
	}

	return nil
}

func (p parser) parseByte(b byte) error {

	// BRACE ESCAPE
	if p.braceLevel != 0 {
		isFirst := p.braceFirst
		p.braceFirst = false

		if b == '{' && !isFirst {
			// increase brace depth
			p.braceLevel++
		} else if b == '}' {
			// decrease brace depth
			p.braceLevel--

			// if this was the last brace, clear the brace escape catch
			if p.braceLevel == 0 {
				p.clearCatch()
			}
		}

		// proceed to the next byte if this was the first or last brace
		if isFirst || p.braceLevel == 0 {
			return nil
		}

		// otherwise, proceed to the catch
		return byteDefault(b)
	}

	// COMMENTS

	// entrance
	if b == '/' && p.next == '*' {
		p.ignoreLevel++

		// this is escaped
		if p.escaped {
			return byteDefault(b)
		}

		// next character
		return nil
	}

	// exit
	if b == '*' && p.next == '/' {

		// we weren't in a comment, so handle normally
		if p.commentLevel == 0 {
			return byteDefault(b)
		}

		// decrease comment level and skip next byte
		p.commentLevel--
		p.skip = true
	}

	// we're inside a comment; skip to next character
	if p.commentLevel != 0 {
		return nil
	}

	// BLOCKS

	if b == '{' {
		// opens a block
		p.ignoreLevel++

		// this is escaped
		if p.escaped {
			return byteDefault(b)
		}

		var blockClasses []string
		var blockType, blockName string

		// if the next char is @, this is {@some_var}
		if p.next == '@' {
			p.skip = true
			blockType = "variable"
		} else {
			var inBlockName, charsScanned int

			// if there is no lastContent, give up because the block has no type
			if len(p.lastContent) == 0 {
				return errors.New("Block has no type")
			}

			// scan the text backward to find the block type and name
			for i := len(p.lastContent) - 1; i != -1; i-- {
				lastChar := p.lastContent[i]
				charsScanned++

				// enter/exit block name
				if lastChar == ']' {
					// entering block name
					inBlockName++

					// we just entered the block name
					if inBlockName == 1 {
						continue
					}
				} else if lastChar == '[' {
					// exiting block name
					inBlockName--

					// we're still in it
					if inBlockName != 0 {
						continue
					}
				}

				// block type/name
				if inBlockName != 0 {
					// we're currently in the block name
					blockName = string(lastChar) + blockName
				} else if matched, _ := regexp.Match(`[\w\-\$\.]`, []byte{b}); matched {
					// this could be part of the block type
					blockType = string(lastChar) + blockType
					continue
				} else if lastChar == '~' && len(blockType) != 0 {
					// tilde terminates block type
					break
				} else if matched, _ := regexp.Match(`\s`, []byte{b}); matched && len(blockType) == 0 {
					// space between things
					continue
				} else {
					// not sure. give this character back and bail
					charsScanned--
					break
				}

				// overwrite last content with the title and name stripped out
				p.lastContent = p.lastContent[:len(p.lastContent)-charsScanned]

				// if the block contains dots, it has classes
				if split := strings.Split(string(blockType), "."); len(split) > 1 {
					blockType, blockClasses = split[0], split[1:]
				}
			}

			// if there is no type at this point, assume it is a map
			if len(blockType) == 0 {
				blockType = "map"
			}

			// if the block type starts with $, it is a model
			if blockType[0] == '$' {
				blockType = blockType[1:]
				blockName = blockType
				blockType = "model"
			}

			// create the block
			block := &parserBlock{
				parser:  &p,
				line:    p.line,
				column:  p.column,
				parent:  p.block,
				typ:     string(blockType),
				name:    string(blockName),
				classes: blockClasses,
				catch:   &parserCatch{},
			}

			// TODO: produce a warning if the block has a name but the type does not support it

			// set the current block
			p.block = block
			p.catch = p.block.catch

			// if the next char is a brace, this is a brace escaped block
			if p.next == '{' {
				p.braceFirst = true
				p.braceLevel++

				// TODO: set the current catch to the brace escape
				// return if catch fails
			}
		}
	} else if b == '}' {
		// closes a block
		p.ignoreLevel++

		// this is escaped
		if p.escaped {
			return byteDefault(b)
		}

		// we cannot close the main block
		if p.block.typ == "main" {
			return errors.New("Attempted to close main block")
		}

		var addContents []interface{}

		// TODO: if/elsif/else statements, {@vars}
		if false {

		} else {
			// normal block. add the block itself
			addContents = []interface{}{p.block}
		}

		// close the block
		p.block.closed = true
		p.block.endLine = p.line
		p.block.endColumn = p.column

		// TODO: clear the catch
		p.catch.appendContent(addContents)

	} else if b == '\\' {
		if p.escaped {
			return byteDefault(b)
		}
		return nil
	}
	return nil
}

func byteDefault(b byte) error {
	return nil
}

func (p parser) clearCatch() {

}
