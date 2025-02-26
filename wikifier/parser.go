package wikifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	wordCharPattern = regexp.MustCompile(`[\w\-\$\.]`)
	spacePattern    = regexp.MustCompile(`\s`)
)

type parser struct {
	pos Position

	last       rune // last rune
	this       rune // current rune
	next       rune // next rune
	next2      rune // next-next rune
	escape     bool // this rune is escaped
	parserChar bool // this character is handled by the master parser (for escapes)
	skip       int  // number of next runes to skip

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
	Line, Column int
}

// none returns true if the position is the default value.
func (pos Position) none() bool {
	return pos.Line == 0 && pos.Column == 0
}

// String returns the position as `{line column}`.
func (pos Position) String() string {
	return fmt.Sprintf("{%d %d}", pos.Line, pos.Column)
}

// MarshalJSON encodes the position to `[line, column]`.
func (pos *Position) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("[%d, %d]", pos.Line, pos.Column)), nil
}

// UnmarshalJSON decodes the position from `[line, column]`.
func (pos *Position) UnmarshalJSON(data []byte) error {
	var val any
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	ary, ok := val.([]any)
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
	pos.Line, pos.Column = int(line), int(col)
	return nil
}

func newParser(page *Page) *parser {
	mb := newBlock("main", "", "", nil, nil, nil, Position{}, page)
	return &parser{block: mb, catch: mb}
}

func (p *parser) parseLine(line []byte, page *Page) error {
	p.pos.Line++
	p.pos.Column = 0 // Reset column position at the start of each line

	// this is a hack to fix extra whitespace in blocks just before they close
	if p.braceLevel == 0 && strings.TrimSpace(string(line)) == "}" {
		line = []byte{'}', '\n'}
	}

	// inject newline back
	if len(line) == 0 || line[len(line)-1] != '\n' {
		line = append(line, '\n')
	}

	// handle each rune
	p.lineHasStarted = false
	for i := 0; i < len(line); {
		r, size := utf8.DecodeRune(line[i:])
		i += size

		// skip this rune
		if p.skip != 0 {
			p.skip--
			continue
		}

		// update column and runes
		p.pos.Column++
		p.this = r

		// next two runes
		if len(line) > i {
			p.next, _ = utf8.DecodeRune(line[i:])
		} else {
			p.next = 0
		}
		if len(line) > i+size {
			p.next2, _ = utf8.DecodeRune(line[i+size:])
		} else {
			p.next2 = 0
		}

		// handle this rune and give up if error occurred
		if err := p.parseRune(r, page); err != nil {
			return err
		}

		// that was the very first non-space character on the line (quiki#3)
		if !p.lineHasStarted && !unicode.IsSpace(r) {
			p.lineHasStarted = true
		}
	}
	return nil
}

// ParserError represents an error in parsing with positional info.
type ParserError struct {
	Pos Position
	Err error
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("{%d %d} %s", e.Pos.Line, e.Pos.Column, e.Err.Error())
}

func (e *ParserError) Unwrap() error {
	return e.Err
}

// creates a ParserError with position and message
func parserError(pos Position, msg string) *ParserError {
	return &ParserError{Pos: pos, Err: errors.New(msg)}
}

var variableTokens = map[rune]bool{
	'@': true,
	'%': true,
	':': true,
	';': true,
	'-': true,
}

func (p *parser) parseRune(r rune, page *Page) error {
	// BRACE ESCAPE
	if p.braceLevel != 0 {
		if r == '{' {
			// increase brace depth
			p.braceLevel++
		} else if r == '}' {
			// decrease brace depth
			p.braceLevel--

			// if this was the last brace, clear the brace escape catch
			if p.braceLevel == 0 {
				p.block.appendContents(p.catch.posContent())
				p.catch = p.catch.parentCatch()
			}
		}

		// proceed to the next rune if this was the first or last brace
		if p.braceLevel == 0 {
			return p.nextRune(r)
		}

		// otherwise, proceed to the catch
		return p.handleRune(r)
	}

	// COMMENTS

	// entrance
	if r == '/' && p.next == '*' {
		p.parserChar = true

		// this is escaped
		if p.escape {
			return p.handleRune(r)
		}

		// next rune
		p.commentLevel++
		return p.nextRune(r)
	}

	// exit
	if r == '*' && p.next == '/' {
		// we weren't in a comment, so handle normally
		if p.commentLevel == 0 {
			return p.handleRune(r)
		}

		// decrease comment level and skip this and next rune
		p.commentLevel--
		p.skip++
		return p.nextRune(r)
	}

	// we're inside a comment; skip to next rune
	if p.commentLevel != 0 {
		return p.nextRune(r)
	}

	// BLOCKS

	if r == '{' {
		// opens a block
		p.parserChar = true

		// this is escaped
		if p.escape {
			return p.handleRune(r)
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
				return errors.New("block has no type")
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
				} else if matched := wordCharPattern.Match([]byte{lastChar}); matched {
					// this could be part of the block type
					blockType = string(lastChar) + blockType
					continue
				} else if lastChar == '~' && len(blockType) != 0 {
					// tilde terminates block type
					break
				} else if matched := spacePattern.Match([]byte{lastChar}); matched && len(blockType) == 0 {
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
			catch := newBraceEscape()
			catch.parent = p.catch
			p.catch = catch
		}

		return p.nextRune(r)
	}

	if r == '}' {
		openPos := p.block.openPosition()

		// closes a block
		p.parserChar = true
		accepting := p.catch.parentCatch()

		// this is escaped
		if p.escape {
			return p.handleRune(r)
		}

		// we cannot close the main block
		if p.block.blockType() == "main" {
			return errors.New("attempted to close main block")
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
				return parserError(openPos, "Unexpected elsif{}")
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
				return parserError(openPos, "Unexpected else{}")
			}

			// title provided
			if p.block.blockName() != "" {
				p.block.warn(openPos, "Condition on else{} ignored")
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
			blk, err := page.GetBlock(varName)
			if err != nil {
				return parserError(openPos, "Variable block @"+varName+" does not contain a block")
			}
			if blk == nil {
				return parserError(openPos, "Variable block @"+varName+" does not exist")
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

		return p.nextRune(r)
	}

	if r == '\\' {
		// the escape will be handled later
		if p.escape {
			return p.handleRune(r)
		}
		return p.nextRune(r)
	}

	// VARIABLES

	// FIXME: these tokens in stray text in the main block cause issues
	if p.block.blockType() == "main" && variableTokens[r] && p.last != '[' {
		// p.parserChar = true ???

		if p.escape {
			return p.handleRune(r)
		}

		// entering a variable declaration on a NEW LINE (quiki#3)
		potentiallyVar := false
		if p.catch == p.block {
			if r == '@' || r == '%' {
				if p.varNegated && p.last == '-' {
					// last char was - for negation, seems likely
					potentiallyVar = true
				} else if !p.lineHasStarted {
					// @ or % started the line, seems likely
					potentiallyVar = true
				}
			} else if r == '-' && !p.lineHasStarted {
				p.varNegated = true
			}
		}

		// ok we're gonna assume it's a variable declaration
		if potentiallyVar {

			// disable interpolation if it's %var
			p.varNotInterpolated = r == '%'

			// catch the var name
			catch := newVariableName(string(r), p.pos)
			catch.parent = p.catch
			p.catch = catch

			return p.nextRune(r)
		}

		// terminate variable name, enter value
		if r == ':' && p.catch.catchType() == catchTypeVariableName {
			// starts a variable value

			// fetch var name, clear the catch
			p.varName = p.catch.lastString()
			p.catch = p.catch.parentCatch()

			// no var name
			if len(p.varName) == 0 {
				return errors.New("variable has no name")
			}

			// now catch the value
			catch := newVariableValue()
			catch.parent = p.catch
			p.catch = catch

			return p.nextRune(r)
		}

		// terminate a boolean
		if r == ';' && p.catch.catchType() == catchTypeVariableName {

			// fetch var name, clear the catch
			p.varName = p.catch.lastString()
			p.catch = p.catch.parentCatch()

			// no var name
			if len(p.varName) == 0 {
				return errors.New("variable has no name")
			}

			// set the value
			page.Set(p.varName, !p.varNegated)

			p.clearVariableState()
			return p.nextRune(r)
		}

		// terminate a string or block variable value
		if r == ';' && p.catch.catchType() == catchTypeVariableValue {

			// we have to also check this here in case it was something like @;
			if len(p.varName) == 0 {
				return errors.New("variable has no name")
			}

			// fetch content and clear catch
			value := fixValuesForStorage(p.catch.content(), p.block, p.pos, !p.varNotInterpolated)
			p.catch = p.catch.parentCatch()

			switch val := value.(type) {
			case []any:
				return fmt.Errorf("variable '%s' contains both text and blocks", p.varName)

			case string, HTML:
				// do nothing
				// (if varNotInterpolated is true, keep string as-is)
				// (if varNotInterpolated is false, would have returned HTML)

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
				return fmt.Errorf("not sure what to do with: %v", val)
			}

			// set the value
			page.Set(p.varName, value)

			p.clearVariableState()
			return p.nextRune(r)
		}

		// negates a boolean variable
		if r == '-' && (p.next == '@' || p.next == '%') {
			// do nothing yet; just make sure we don't get to default
			return p.nextRune(r)
		}

		// default
		return p.handleRune(r)
	}

	return p.handleRune(r)
}

// (NEXT DEFAULT)
func (p *parser) handleRune(r rune) error {

	// if we have someplace to append this content, do that
	if p.catch == nil {
		// nothing to catch! I don't think this can ever happen since the main block
		// is the top-level catch and cannot be closed, but it's here just in case
		return errors.New("Nothing to catch rune: " + string(r))
	}

	// at this point, anything that needs escaping should have been handled.
	// so, if this rune is escaped and reached all the way to here, we will
	// pretend it's not escaped by reinjecting a backslash. this allows
	// further parsers to handle escapes (in particular, Formatter.)
	add := string(r)
	if p.escape && !p.parserChar {
		add = string([]rune{p.last, r})
	}

	// terminate the catch if the catch says to skip this rune
	if p.catch.shouldSkipRune(r) {

		// fetch the stuff caught up to this point
		pc := p.catch.posContent()

		// also, fetch prefixes if there are any
		if pfx := p.catch.positionedPrefixContent(); pfx != nil {
			pc = append(pfx, pc...)
		}

		// revert to the parent catch, and add our stuff to it
		p.catch = p.catch.parentCatch()
		p.catch.appendContents(pc)

	} else if !p.catch.runeOk(r) {
		// ask the catch if this rune is acceptable

		char := string(r)
		if char == "\n" {
			char = "\u2424"
		}
		err := "Invalid rune '" + char + "' in " + string(p.catch.catchType()) + "."
		if str := p.catch.lastString(); str != "" {
			err += " Partial: " + str
		}
		return errors.New(err)
	}

	// so um, if the content is whitespace/newline
	// and the catch has no content yet, ignore this
	if len(p.catch.content()) == 0 && (r == '\n') {
		return p.nextRune(r)
	}

	// append
	p.catch.appendContent(add, p.pos)

	return p.nextRune(r)
}

// (NEXT RUNE)
func (p *parser) nextRune(r rune) error {

	// if current rune is \, set escape for the next
	if r == '\\' && !p.escape && p.braceLevel == 0 {
		p.escape = true
	} else {
		p.escape = false
	}

	p.parserChar = false
	p.last = r
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
	blk.warn(blk.openPosition(), "Invalid "+blk.blockType()+"{} condition; expected variable or attribute")
	return false
}

func (p *parser) clearVariableState() {
	p.varName = ""
	p.varNotInterpolated = false
	p.varNegated = false
}
