package wikifier

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

type parser struct {
	pos parserPosition

	last   byte // last byte
	this   byte // current byte
	next   byte // next byte
	skip   bool // skip next byte
	ignore bool // this byte is ignored
	escape bool // this byte is escaped

	catch parserCatch
	block *parserBlock

	commentLevel int
	braceLevel   int
	braceFirst   bool
}

type parserPosition struct {
	line, column int
}

func Parse(input string) error {
	mb := &parserBlock{typ: "main", genericCatch: &genericCatch{}}
	p := &parser{block: mb}
	p.catch = mb

	for i, line := range strings.Split(input, "\n") {
		p.pos.line = i + 1
		if err := p.parseLine([]byte(line)); err != nil {
			return err
		}
	}

	log.Printf("OK so at the end, main block content is %+v", mb.getContent())
	return nil
}

func (p *parser) parseLine(line []byte) error {
	for i, b := range line {

		// skip this byte
		if p.skip {
			p.skip = false
			continue
		}

		// update column and bytes
		p.pos.column = i + 1
		p.this = b

		if len(line) > i+1 {
			p.next = line[i+1]
		} else {
			p.next = 0
		}

		// handle this byte and give up if error occurred
		if err := p.parseByte(b); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) parseByte(b byte) error {
	log.Printf("parseByte(%s, last: %s, next: %s)", string(b), string(p.last), string(p.next))

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
				p.catch = p.catch.getParent()
			}
		}

		// proceed to the next byte if this was the first or last brace
		if isFirst || p.braceLevel == 0 {
			return p.nextByte(b)
		}

		// otherwise, proceed to the catch
		return p.handleByte(b)
	}

	// COMMENTS

	// entrance
	if b == '/' && p.next == '*' {
		p.ignore = true

		// this is escaped
		if p.escape {
			return p.handleByte(b)
		}

		// next byte
		p.commentLevel++
		log.Println("increased comment level to", p.commentLevel)
		return p.nextByte(b)
	}

	// exit
	if b == '*' && p.next == '/' {

		// we weren't in a comment, so handle normally
		if p.commentLevel == 0 {
			return p.handleByte(b)
		}

		// decrease comment level and skip this and next byte
		p.commentLevel--
		log.Println("decreased comment level to", p.commentLevel)
		p.skip = true
		return p.nextByte(b)
	}

	// we're inside a comment; skip to next byte
	if p.commentLevel != 0 {
		return p.nextByte(b)
	}

	// BLOCKS

	if b == '{' {
		// opens a block
		p.ignore = true

		// this is escaped
		if p.escape {
			return p.handleByte(b)
		}

		var blockClasses []string
		var blockType, blockName string

		// if the next char is @, this is {@some_var}
		if p.next == '@' {
			p.skip = true
			blockType = "variable"
		} else {
			var inBlockName, charsScanned int
			lastContent := p.catch.getLastString()
			log.Printf("LAST CONTENT: %v", lastContent)

			// if there is no lastContent, give up because the block has no type
			if len(lastContent) == 0 {
				return errors.New("Block has no type")
			}

			// scan the text backward to find the block type and name
			for i := len(lastContent) - 1; i != -1; i-- {
				lastChar := lastContent[i]
				charsScanned++

				// enter/exit block name
				if lastChar == ']' {
					log.Println("entering block name")
					// entering block name
					inBlockName++

					// we just entered the block name
					if inBlockName == 1 {
						continue
					}
				} else if lastChar == '[' {
					log.Println("exiting block name")

					// exiting block name
					inBlockName--

					// we're still in it
					if inBlockName != 1 {
						continue
					}
				}

				// block type/name
				if inBlockName != 0 {
					// we're currently in the block name
					blockName = string(lastChar) + blockName
				} else if matched, _ := regexp.Match(`[\w\-\$\.]`, []byte{lastChar}); matched {
					// this could be part of the block type
					blockType = string(lastChar) + blockType
					continue
				} else if lastChar == '~' && len(blockType) != 0 {
					// tilde terminates block type
					break
				} else if matched, _ := regexp.Match(`\s`, []byte{lastChar}); matched && len(blockType) == 0 {
					// space between things
					continue
				} else {
					// not sure. give this byte back and bail
					log.Printf("giving up due to: %v", string(lastChar))
					charsScanned--
					break
				}
			}

			// overwrite last content with the title and name stripped out
			log.Printf("Setting last content to: %v", lastContent[:len(lastContent)-charsScanned])
			p.catch.setLastContent(lastContent[:len(lastContent)-charsScanned])

			// if the block contains dots, it has classes
			if split := strings.Split(string(blockType), "."); len(split) > 1 {
				blockType, blockClasses = split[0], split[1:]
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
			log.Printf("Creating block: %s[%s]{}", blockType, blockName)
			block := &parserBlock{
				parser:       p,
				openPos:      p.pos,
				parent:       p.block,
				typ:          blockType,
				name:         blockName,
				classes:      blockClasses,
				genericCatch: &genericCatch{},
			}

			// TODO: produce a warning if the block has a name but the type does not support it

			// set the current block
			p.block = block
			p.catch = block

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
		p.ignore = true

		// this is escaped
		if p.escape {
			return p.handleByte(b)
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
		p.block.closePos = p.pos

		// clear the catch
		p.block = p.block.parent
		p.catch = p.catch.getParent()
		p.catch.appendContent(addContents, p.pos)

	} else if b == '\\' {
		if p.escape {
			return p.handleByte(b)
		}
		return p.nextByte(b)
	}
	return p.handleByte(b)
}

// (NEXT DEFAULT)
func (p *parser) handleByte(b byte) error {
	log.Println("handleByte", string(b))

	// if we have someplace to append this content, do that
	if p.catch == nil {
		// nothing to catch! I don't think this can ever happen since the main block
		// is the top-level catch and cannot be closed, but it's here just in case
		return errors.New("Nothing to catch byte: " + string(b))
	}

	// at this point, anything that needs escaping should have been handled.
	// so, if this byte is escaped and reached all the way to here, we will
	// pretend it's not escaped by reinjecting a backslash. this allows
	// further parsers to handle escapes (in particular, Formatter.)
	add := string(b)
	if p.ignore && p.escape {
		add = string([]byte{p.last, b})
	}

	// terminate the catch if the byte is in those to skip
	if p.catch.shouldSkipByte(b) {

		// // fetch the stuff caught up to this point
		// content := p.catch.getContent()
		// position := p.catch.getPosition()

		// // also, fetch prefixes if there are any
		// if p.catch.getPrefixContent() != nil {
		// 	content = append([]interface{}{p.catch.getPrefixContent()}, content...)
		// 	position = append(p.catch.getPrefixPosition(), position...)
		// }

		// // revert to the parent catch, and add our stuff to it
		// p.catch = p.catch.getParent()
		// p.catch.pushContent(content, position)
	}

	// FIXME: make sure the char is acceptable according to valid chars
	//     # make sure the char is acceptable according to valid_chars
	//     elsif ($catch->{valid_chars} && $char !~ $catch->{valid_chars}) {
	//         my $loc = $catch->{location}[-1];
	//         $char   = "\x{2424}" if $char eq "\n";
	//         my $err = "Invalid byte '$char' in $$catch{hr_name}.";
	//         $err   .= " Partial: $loc" if length $loc;
	//         return $c->error($err);
	//     }

	// append
	p.catch.appendString(add, p.pos)

	return p.nextByte(b)
}

// (NEXT BYTE)
func (p *parser) nextByte(b byte) error {
	log.Println("nextByte", string(b))

	p.ignore = false
	p.last = b

	// if current byte is \, set escape for the next
	if b == '\\' && !p.escape && p.braceLevel == 0 {
		p.escape = true
	} else {
		p.escape = false
	}

	return nil
}
