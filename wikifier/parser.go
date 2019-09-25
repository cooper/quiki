package wikifier

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type parser struct {
	pos position

	last   byte // last byte
	this   byte // current byte
	next   byte // next byte
	skip   bool // skip next byte
	escape bool // this byte is escaped

	catch catch // current parser catch
	block block // current parser block

	commentLevel int  // comment depth
	braceLevel   int  // brace escape depth
	braceFirst   bool // true when entering a brace escape

	varName            string
	varNotInterpolated bool
	varNegated         bool

	conditional       bool // current conditional
	conditionalExists bool
}

type position struct {
	line, column int
}

var variableTokens = map[byte]bool{
	'@': true,
	'%': true,
	':': true,
	';': true,
	'-': true,
}

func (pos position) none() bool {
	return pos.line == 0 && pos.column == 0
}

func newParser() *parser {
	mb := newBlock("main", "", nil, nil, nil, position{})
	return &parser{block: mb, catch: mb}
}

func (p *parser) parseLine(line []byte, page *Page) error {

	// inject newline back
	if len(line) == 0 || line[len(line)-1] != '\n' {
		line = append(line, '\n')
	}

	// handle each byte
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
		if err := p.parseByte(b, page); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseByte(b byte, page *Page) error {
	log.Printf("parseByte(%s, last: %s, next: %s)", string(b), string(p.last), string(p.next))

	// fix extra newline added to code{} blocks
	if b == '{' && p.next == '\n' {
		p.skip = true
	}

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
				p.block.appendContents(p.catch.posContent())
				p.catch = p.catch.parentCatch()
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
			lastContent := p.catch.lastString()
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
		}

		// if there is a name but no type, it's a section with a heading
		// if neither, it's a map
		if len(blockType) == 0 {
			if len(blockName) != 0 {
				blockType = "sec"
			} else {
				blockType = "map"
			}
		}

		// if the block type starts with $, it is a model
		if blockType[0] == '$' {
			blockType = blockType[1:]
			blockName = blockType
			blockType = "model"
		}

		// create the block
		log.Printf("Creating block: %s[%s]{}", blockType, blockName)
		block := newBlock(blockType, blockName, blockClasses, p.block, p.catch, p.pos)

		// TODO: produce a warning if the block has a name but the type does not support it

		// set the current block
		p.block = block
		p.catch = block

		// if the next char is a brace, this is a brace escaped block
		if p.next == '{' {
			p.braceFirst = true
			p.braceLevel++

			// start the brace escape catch
			catch := newBraceEscape(p.pos)
			catch.parent = p.catch
			p.catch = catch
		}

		return p.nextByte(b)
	}

	if b == '}' {
		// closes a block
		accepting := p.catch.parentCatch()

		// this is escaped
		if p.escape {
			return p.handleByte(b)
		}

		// we cannot close the main block
		if p.block.blockType() == "main" {
			return errors.New("Attempted to close main block")
		}

		// if{}, elsif{}, else{}, {@vars}
		switch p.block.blockType() {

		// if [condition] { ... }
		case "if":

			p.conditionalExists = true
			p.conditional = p.getConditional(page, p.block.blockName())
			if p.conditional {
				accepting.appendContents(p.block.posContent())
			}

		// elsif [condition] { ... }
		case "elsif":

			// no conditional exists before this
			if !p.conditionalExists {
				return errors.New("Unexpected elsif{}")
			}

			// only evaluate the conditional if the last one was false
			if !p.conditional {
				p.conditional = p.getConditional(page, p.block.blockName())
				if p.conditional {
					accepting.appendContents(p.block.posContent())
				}
			}

		// else { ... }
		case "else":

			// no conditional exists before this
			if !p.conditionalExists {
				return errors.New("Unexpected elsif{}")
			}

			// title provided
			if p.block.blockName() != "" {
				p.block.warn(p.pos, "Conditional on else{} ignored")
			}

			// the condition was false. add the contents of the else.
			if !p.conditional {
				accepting.appendContents(p.block.posContent())
			}

			// reset the conditional
			p.conditionalExists = false

		// {@var}
		case "variable":

			// fetch variable name
			varName := p.block.blockName()
			if varName == "" {
				// consider: clear the block content? does it matter?
				varName = p.block.lastString()
			}

			// find the value and make sure it's a block
			obj, err := page.GetBlock(varName)
			if err != nil {
				return errors.New("Variable block @" + varName + " does not contain a block")
			}
			if obj == nil {
				return errors.New("Variable block @" + varName + " does not exist")
			}
			blk, ok := obj.(block)
			if !ok {
				return errors.New("Variable block @" + varName + " does not contain a block")
			}

			// overwrite the block's parent to the parent of the {@var}
			blk.setParentBlock(p.block.parentBlock())

			// add this block
			accepting.appendContent(blk, p.pos)

		// normal block. add the block itself
		default:
			// FIXME: this is disabled for now because it causes problems when
			// there are blocks inside of else{}
			// p.conditionalExists = false
			accepting.appendContent(p.block, p.pos)
		}

		// close the block
		p.block.close(p.pos)
		p.block = p.block.parentBlock()
		p.catch = p.catch.parentCatch()

		return p.nextByte(b)
	}

	if b == '\\' {
		// the escape will be handled later
		if p.escape {
			return p.handleByte(b)
		}
		return p.nextByte(b)
	}

	// VARIABLES

	if p.block.blockType() == "main" && variableTokens[b] && p.last != '[' {
		log.Println("variable tok", string(b))

		if p.escape {
			return p.handleByte(b)
		}

		// entering a variable declaration
		if (b == '@' || b == '%') && p.catch == p.block {
			log.Println("variable declaration", string(b))

			// disable interpolation if it's %var
			if b == '%' {
				p.varNotInterpolated = true
			}

			// negate the value if -@var
			pfx := string(b)
			if p.last == '-' {
				pfx = string(p.last) + pfx
				p.varNegated = true
			}

			// catch the var name
			catch := newVariableName(pfx, p.pos)
			catch.parent = p.catch
			p.catch = catch

			return p.nextByte(b)
		}

		// terminate variable name, enter value
		if b == ':' && p.catch.catchType() == catchTypeVariableName {
			// starts a variable value

			// fetch var name, clear the catch
			p.varName = p.catch.lastString()
			p.catch = p.catch.parentCatch()

			// no var name
			if len(p.varName) == 0 {
				return errors.New("Variable has no name")
			}
			log.Printf("VALUE VAR NAME: %v", p.varName)

			// now catch the value
			catch := newVariableValue()
			catch.parent = p.catch
			p.catch = catch

			log.Printf("SETTING CATCH FOR VAR: %+v", p.catch)

			return p.nextByte(b)
		}

		// terminate a boolean
		if b == ';' && p.catch.catchType() == catchTypeVariableName {

			// fetch var name, clear the catch
			p.varName = p.catch.lastString()
			p.catch = p.catch.parentCatch()

			// no var name
			if len(p.varName) == 0 {
				return errors.New("Variable has no name")
			}
			log.Printf("BOOLEAN VAR NAME: %v", p.varName)

			// set the value
			page.Set(p.varName, !p.varNegated)

			p.clearVariableState()
			return p.nextByte(b)
		}

		// terminate a string or block variable value
		if b == ';' && p.catch.catchType() == catchTypeVariableValue {

			// we have to also check this here in case it was something like @;
			if len(p.varName) == 0 {
				return errors.New("Variable has no name")
			}

			// fetch content and clear catch
			value := fixValuesForStorage(p.catch.content())
			p.catch = p.catch.parentCatch()

			log.Println("Variable values = ", value)

			switch val := value.(type) {
			case []interface{}:
				return fmt.Errorf("Variable '%s' contains both text and blocks", p.varName)

			case string:
				log.Println("Got var str:", val)

				// format it unless told not to
				if !p.varNotInterpolated {
					value = page.parseFormattedText(val)
				}

			case block:

				// parse the block
				// note: this means the block will be parsed twice
				// once now so that vars/warnings can be produced
				// once later when it is injected..
				// just cuz there is no way to tell that it has been done already
				val.parse(page)
				log.Println("Got var block:", val)

			case nil:
				return fmt.Errorf("there's nothing here")

			default:
				return fmt.Errorf("Not sure what to do with: %v", val)
			}

			// set the value
			log.Println("Setting", p.varName, value)
			page.Set(p.varName, value)

			p.clearVariableState()
			return p.nextByte(b)
		}

		// negates a boolean variable
		if b == '-' && (p.next == '@' || p.next == '%') {
			// do nothing yet; just make sure we don't get to default
			return p.nextByte(b)
		}

		log.Printf("MADE IT DOWN WITH %v; CATCH: %v", string(b), p.catch)

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
	if p.escape {
		add = string([]byte{p.last, b})
	}

	// terminate the catch if the catch says to skip this byte
	if p.catch.shouldSkipByte(b) {

		// fetch the stuff caught up to this point
		pc := p.catch.posContent()

		// also, fetch prefixes if there are any
		if pfx := p.catch.positionedPrefixContent(); pfx != nil {
			pc = append(pfx, pc...)
		}

		// revert to the parent catch, and add our stuff to it
		p.catch = p.catch.parentCatch()
		p.catch.appendContents(pc)

	} else if !p.catch.byteOK(b) {
		// ask the catch if this byte is acceptable

		char := string(b)
		if char == "\n" {
			char = "\u2424"
		}
		err := "Invalid byte '" + char + "' in " + string(p.catch.catchType()) + "."
		if str := p.catch.lastString(); str != "" {
			err += "Partial: " + str
		}
		return errors.New(err)
	}

	// append
	p.catch.appendContent(add, p.pos)

	return p.nextByte(b)
}

// (NEXT BYTE)
func (p *parser) nextByte(b byte) error {
	log.Println("nextByte", string(b))

	p.last = b

	// if current byte is \, set escape for the next
	if b == '\\' && !p.escape && p.braceLevel == 0 {
		p.escape = true
	} else {
		p.escape = false
	}

	return nil
}

func (p *parser) getConditional(page *Page, condition string) bool {

	// no condition
	if condition == "" {
		p.warn("Conditional has no condition")
	}

	// negated
	if condition[0] == '!' {
		return !p.getConditional(page, condition[1:])
	}

	// looks like a variable
	if condition[0] == '@' {

		// check boolean first
		b, err1 := page.GetBool(condition[1:])

		// possibly just says this is not a boolean
		if err1 != nil {

			// try a generic lookup
			v, err2 := page.Get(condition[1:])

			// still error? this is something serious
			if err2 != nil {
				p.warn(err2.Error())
				return false
			}

			// something's there
			return v != nil
		}

		return b
	}

	// something else
	p.warn("Invalid condition; expected variable or attribute")
	return false
}

func (p *parser) clearVariableState() {
	p.varName = ""
	p.varNotInterpolated = false
	p.varNegated = false
}

func (p *parser) warn(warning string) {
	log.Printf("WARNING: at %v: %s", p.pos, warning)
}
