package wikifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type parser struct {
	pos Position

	last       byte // last byte
	this       byte // current byte
	next       byte // next byte
	next2      byte // next-next byte
	escape     bool // this byte is escaped
	parserChar bool // this character is handled by the master parser (for escapes)
	skip       int  // number of next bytes to skip

	catch catch // current parser catch
	block block // current parser block

	commentLevel int // comment depth
	braceLevel   int // brace escape depth

	varName            string
	varNotInterpolated bool
	varNegated         bool

	conditional       bool // current conditional
	conditionalExists bool

	lineHasStarted bool // true once the first non-space has occurred
}

// Position represents a line and column position within a quiki source file.
type Position struct {
	line, column int
}

var variableTokens = map[byte]bool{
	'@': true,
	'%': true,
	':': true,
	';': true,
	'-': true,
}

func (pos Position) none() bool {
	return pos.line == 0 && pos.column == 0
}

// String returns the position as `{line column}`.
func (pos Position) String() string {
	return fmt.Sprintf("{%d %d}", pos.line, pos.column)
}

// MarshalJSON encodes the position to `[line, column]`.
func (pos Position) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("[%d, %d]", pos.line, pos.column)), nil
}

// UnmarshalJSON decodes the position from `[line, column]`.
func (pos Position) UnmarshalJSON(data []byte) error {
	var val interface{}
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	ary, ok := val.([]interface{})
	if !ok || len(ary) != 2 {
		return errors.New("(Position).UnmarshalJSON: expected JSON array with len(2)")
	}
	line, ok := ary[0].(float64)
	if !ok {
		return errors.New("(Position).UnmarshalJSON: expected line number to be integer")
	}
	col, ok := ary[1].(float64)
	if !ok {
		return errors.New("(Position).UnmarshalJSON: expected column number to be integer")
	}
	pos.line, pos.column = int(line), int(col)
	return nil
}

func newParser(page *Page) *parser {
	mb := newBlock("main", "", "", nil, nil, nil, Position{}, page)
	return &parser{block: mb, catch: mb}
}

func (p *parser) parseLine(line []byte, page *Page) error {
	p.pos.line++

	// this is a hack to fix extra whitespace in blocks just before they close
	if p.braceLevel == 0 && strings.TrimSpace(string(line)) == "}" {
		line = []byte{'}', '\n'}
	}

	// inject newline back
	if len(line) == 0 || line[len(line)-1] != '\n' {
		line = append(line, '\n')
	}

	// handle each byte
	p.lineHasStarted = false
	for i, b := range line {

		// skip this byte
		if p.skip != 0 {
			p.skip--
			continue
		}

		// update column and bytes
		p.pos.column = i + 1
		p.this = b

		// next two bytes
		if len(line) > i+1 {
			p.next = line[i+1]
		} else {
			p.next = 0
		}
		if len(line) > i+2 {
			p.next2 = line[i+2]
		} else {
			p.next2 = 0
		}

		// handle this byte and give up if error occurred
		if err := p.parseByte(b, page); err != nil {
			return err
		}

		// that was the very first non-space character on the line (quiki#3)
		if !p.lineHasStarted && !unicode.IsSpace(rune(b)) {
			p.lineHasStarted = true
		}
	}
	return nil
}

func (p *parser) parseByte(b byte, page *Page) error {

	// // fix extra newline added to code{} blocks
	// if p.braceLevel == 0 && b == '{' && p.next == '\n' {
	// 	p.skip++
	// }

	// BRACE ESCAPE
	if p.braceLevel != 0 {

		if b == '{' {
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
		if p.braceLevel == 0 {
			return p.nextByte(b)
		}

		// otherwise, proceed to the catch
		return p.handleByte(b)
	}

	// COMMENTS

	// entrance
	if b == '/' && p.next == '*' {
		p.parserChar = true

		// this is escaped
		if p.escape {
			return p.handleByte(b)
		}

		// next byte
		p.commentLevel++
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
		p.skip++
		return p.nextByte(b)
	}

	// we're inside a comment; skip to next byte
	if p.commentLevel != 0 {
		return p.nextByte(b)
	}

	// BLOCKS

	if b == '{' {
		// opens a block
		p.parserChar = true

		// this is escaped
		if p.escape {
			return p.handleByte(b)
		}

		var blockClasses []string
		var blockType, blockName, headingID string
		var inHeadingID bool

		// if the next char is @, this is {@some_var}
		if p.next == '@' {
			p.skip++
			blockType = "variable"
		} else {
			var inBlockName, charsScanned int
			lastContent := p.catch.lastString()

			// if there is no lastContent, give up because the block has no type
			if len(lastContent) == 0 {
				return errors.New("Block has no type")
			}

			// scan the text backward to find the block type and name
			for i := len(lastContent) - 1; i != -1; i-- {
				lastChar := lastContent[i]
				charsScanned++

				// enter/exit block name or heading ID
				if lastChar == ']' && blockType == "" {
					// entering block name
					inBlockName++

					// we just entered the block name
					if inBlockName == 1 {
						inHeadingID = false
						continue
					}
				} else if lastChar == '[' {

					// exiting block name
					inBlockName--

					// we're still in it
					if inBlockName != 1 {
						continue
					}
				} else if lastChar == '#' && blockName == "" && blockType == "" {
					// enter/exit heading ID
					inHeadingID = headingID == ""
					continue
				}

				// block type/name
				if inBlockName != 0 {
					// we're currently in the block name
					blockName = string(lastChar) + blockName
				} else if inHeadingID {
					// we're currently in the heading ID
					if lastChar != ' ' && lastChar != '\t' {
						headingID = string(lastChar) + headingID
					}
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
					charsScanned--
					break
				}
			}

			// overwrite last content with the title and name stripped out
			p.catch.setLastContent(lastContent[:len(lastContent)-charsScanned])

			// if the block contains dots, it has classes
			if split := strings.Split(string(blockType), "."); len(split) > 1 {
				blockType, blockClasses = split[0], split[1:]
			}
		}

		// if the current block is an infobox{}, sub-blocks are infosec{}
		// otherwise:
		// if there is a name but no type, it's a section with a heading
		// if neither, it's a map
		if len(blockType) == 0 {
			if p.block.blockType() == "infobox" {
				blockType = "infosec"
			} else if len(blockName) != 0 {
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
		block := newBlock(blockType, blockName, headingID, blockClasses, p.block, p.catch, p.pos, page)

		// TODO: produce a warning if the block has a name but the type does not support it

		// set the current block
		p.block = block
		p.catch = block

		// if the next char is a brace, this is a brace escaped block
		if p.next == '{' {
			p.braceLevel++

			// skip the next brace
			p.skip++

			// also skip the newline after it
			if p.next2 == '\n' {
				p.skip++
			}

			// start the brace escape catch
			catch := newBraceEscape(p.pos)
			catch.parent = p.catch
			p.catch = catch
		}

		return p.nextByte(b)
	}

	if b == '}' {
		// closes a block
		p.parserChar = true
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
			p.conditional = p.getConditional(p.block, page, p.block.blockName())
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
				p.conditional = p.getConditional(p.block, page, p.block.blockName())
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

	// FIXME: these tokens in stray text in the main block cause issues
	if p.block.blockType() == "main" && variableTokens[b] && p.last != '[' {
		// p.parserChar = true ???

		if p.escape {
			return p.handleByte(b)
		}

		// entering a variable declaration on a NEW LINE (quiki#3)
		potentiallyVar := false
		if p.catch == p.block {
			if b == '@' || b == '%' {
				if p.varNegated && p.last == '-' {
					// last char was - for negation, seems likely
					potentiallyVar = true
				} else if !p.lineHasStarted {
					// @ or % started the line, seems likely
					potentiallyVar = true
				}
			} else if b == '-' && !p.lineHasStarted {
				p.varNegated = true
			}
		}

		// ok we're gonna assume it's a variable declaration
		if potentiallyVar {

			// disable interpolation if it's %var
			if b == '%' {
				p.varNotInterpolated = true
			}

			// catch the var name
			catch := newVariableName(string(b), p.pos)
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

			// now catch the value
			catch := newVariableValue()
			catch.parent = p.catch
			p.catch = catch

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
			value := fixValuesForStorage(p.catch.content(), page, true)
			p.catch = p.catch.parentCatch()

			switch val := value.(type) {
			case []interface{}:
				return fmt.Errorf("Variable '%s' contains both text and blocks", p.varName)

			case string:

				// replace newlines with spaces
				val = strings.ReplaceAll(val, "\n", " ")

				// format it unless told not to
				if !p.varNotInterpolated {
					value = page.formatText(val)
				}

			case HTML:
				value = val

			case block:

				// parse the block
				// note: this means the block will be parsed twice
				// once now so that vars/warnings can be produced
				// once later when it is injected..
				// just cuz there is no way to tell that it has been done already
				val.parse(page)

			case nil:
				// empty string
				value = ""

			default:
				return fmt.Errorf("Not sure what to do with: %v", val)
			}

			// set the value
			page.Set(p.varName, value)

			p.clearVariableState()
			return p.nextByte(b)
		}

		// negates a boolean variable
		if b == '-' && (p.next == '@' || p.next == '%') {
			// do nothing yet; just make sure we don't get to default
			return p.nextByte(b)
		}

		// default
		return p.handleByte(b)
	}

	return p.handleByte(b)
}

// (NEXT DEFAULT)
func (p *parser) handleByte(b byte) error {

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
	if p.escape && !p.parserChar {
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
			err += " Partial: " + str
		}
		return errors.New(err)
	}

	// so um, if the content is whitespace/newline
	// and the catch has no content yet, ignore this
	if len(p.catch.content()) == 0 && (b == '\n') {
		return p.nextByte(b)
	}

	// append
	p.catch.appendContent(add, p.pos)

	return p.nextByte(b)
}

// (NEXT BYTE)
func (p *parser) nextByte(b byte) error {

	// if current byte is \, set escape for the next
	if b == '\\' && !p.escape && p.braceLevel == 0 {
		p.escape = true
	} else {
		p.escape = false
	}

	p.parserChar = false
	p.last = b
	return nil
}

func (p *parser) getConditional(blk block, page *Page, condition string) bool {

	// no condition
	if condition == "" {
		blk.warn(blk.openPosition(), "Conditional "+blk.blockType()+"{} has no condition")
		return false
	}

	// negated
	if condition[0] == '!' {
		return !p.getConditional(blk, page, condition[1:])
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
				blk.warn(p.pos, err2.Error())
				return false
			}

			// something's there
			return v != nil
		}

		return b
	}

	// something else
	blk.warn(blk.openPosition(), "Invalid condition; expected variable or attribute")
	return false
}

func (p *parser) clearVariableState() {
	p.varName = ""
	p.varNotInterpolated = false
	p.varNegated = false
}
