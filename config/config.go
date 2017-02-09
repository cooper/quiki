package config

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

// buffer types
const (
	NO_BUF    = iota // no buffer
	VAR_NAME  = iota // variable name
	VAR_VALUE = iota // string variable value
)

type Config struct {
	path string
	vars map[string]interface{}
}

// parser state info
type parserState struct {
	lastByte     byte          // previous byte
	commentLevel uint8         // comment level
	inComment    bool          // true if in a comment
	escaped      bool          // true if the current byte is escaped
	buffer       *bytes.Buffer // current buffer
	buffType     uint8         // type of buffer
	varName      string        // current variable name
	varPercent   bool          // true if current variable is a %var
}

// increase block comment level
func (state *parserState) increaseCommentLevel() {
	state.removeLastNBytes(1)
	state.commentLevel++
	state.inComment = true
}

// decrease block comment level
func (state *parserState) decreaseCommentLevel() {
	if state.commentLevel <= 0 {
		return
	}
	state.commentLevel--
	state.inComment = state.commentLevel != 0
}

// remove the last N bytes from the buffer
func (state *parserState) removeLastNBytes(n int) {
	if state.buffer == nil || state.buffer.Len() < n {
		return
	}
	state.buffer.Truncate(state.buffer.Len() - n)
}

// start a new buffer
func (state *parserState) startBuffer(t uint8) {
	state.buffer = new(bytes.Buffer)
	state.buffType = t
}

// destroy a buffer, returning its contents
func (state *parserState) endBuffer() string {
	str := state.buffer.String()
	state.buffer = nil
	state.buffType = NO_BUF
	return str
}

// destroy the current variable state, returning the variable name
func (state *parserState) getVariable() string {
	name := state.varName
	state.varName = ""
	state.varPercent = false
	return name
}

// new config
func New(path string) *Config {
	return &Config{
		path: path,
		vars: make(map[string]interface{}),
	}
}

// parse config
func (conf *Config) Parse() error {

	// open the config
	file, err := os.Open(conf.path)
	if err != nil {
		return err
	}
	defer file.Close()

	state := &parserState{}
	for {
		b := make([]byte, 1)
		_, err := file.Read(b)

		// eof
		if err == io.EOF {
			break
		}

		// some other error
		if err != nil {
			return err
		}

		// handle the character
		err = conf.handleByte(state, b[0])
		state.lastByte = b[0]

		// acter error
		if err != nil {
			return err
		}
	}

	return nil
}

// handle one byte
func (conf *Config) handleByte(state *parserState, b byte) error {

	// this character is escaped
	if state.escaped {
		b = 0
		state.escaped = false
	}

	if state.lastByte == '/' && b == '*' {

		// comment entrance
		state.increaseCommentLevel()
		return nil

	} else if state.lastByte == '*' && b == '/' {

		// comment closure
		state.decreaseCommentLevel()
		return nil

	} else if state.inComment {

		// we're in a comment currently
		return nil
	}

	switch b {

	// escape
	case '\\':
		state.escaped = true

		// start of a variable
	case '@', '%':

		// we're already in a variable name
		if state.buffType == VAR_NAME {
			return errors.New("Already in variable name @" + state.buffer.String())
		}

		// we're not in a value, so this starts a variable
		if state.buffType != NO_BUF {
			goto realDefault
		}
		state.varPercent = b == '%'
		state.startBuffer(VAR_NAME)

	// end of variable name
	case ':':

		// we're in the var name, so terminate it
		if state.buffType != VAR_NAME {
			goto realDefault
		}
		state.varName = state.endBuffer()
		state.startBuffer(VAR_VALUE)

	// end of a variable definition
	case ';':

		if state.buffType == VAR_NAME {

			// terminating a boolean
			state.endBuffer()
			conf.Set(state.getVariable(), "1")

		} else if state.buffType == VAR_VALUE {

			// terminating a string
			value := conf.parseFormatting(state, state.endBuffer())
			conf.Set(state.getVariable(), value)

		} else {
			goto realDefault
		}
		break

	default:
		goto realDefault
	}

	// this is skipped if going to realDefault
	return nil

realDefault:

	// we're in a comment; ignore this
	if state.inComment {
		return nil
	}

	// otherwise, write this to the current buffer
	if state.buffer != nil {
		state.buffer.WriteByte(b)
	}

	return nil
}

// return the map and attribute name for a variable name
func (conf *Config) getWhere(varName string, createAsNeeded bool) (map[string]interface{}, string) {

	// split up into parts
	var parts = strings.Split(varName, ".")
	if len(parts) == 0 {
		return nil, ""
	}

	// the last one is the final variable name
	lastPart, parts := parts[len(parts)-1], parts[:len(parts)-1]

	// start with the main map
	where := conf.vars

	// for each part, fetch the map inside
	for _, part := range parts {

		// find interface
		iface := where[part]
		if iface == nil {
			if !createAsNeeded {
				return nil, ""
			}

			// maybe create a map
			iface = make(map[string]interface{})
			where[part] = iface
		}

		// find map
		switch aMap := iface.(type) {
		case map[string]interface{}:
			where = aMap

		// nothing there; give up
		default:
			return nil, ""
		}
	}

	return where, lastPart
}

// parse formatted text
func (conf *Config) parseFormatting(state *parserState, format string) string {

	// trim whitespace before anything else
	format = strings.TrimSpace(format)

	// this is a %var, so we shouldn't format it
	if state.varPercent {
		return format
	}

	return "TODO"
}

// get string value
func (conf *Config) Get(varName string) string {

	// get the map
	where, lastPart := conf.getWhere(varName, false)
	if where == nil {
		log.Println("config: Could not Get @" + varName)
		return ""
	}

	// get the string value
	iface := where[lastPart]
	switch str := iface.(type) {
	case string:
		return str
	}

	log.Println("config: @" + varName + "is not a string")
	return ""
}

func (conf *Config) Set(varName string, value string) {

	// get the map
	where, lastPart := conf.getWhere(varName, true)
	if where == nil {
		log.Println("config: Could not Set @" + varName)
		return
	}

	// set the string value
	where[lastPart] = value
}

// get bool value
func (conf *Config) GetBool(varName string) bool {
	return isTrueString(conf.Get(varName))
}

// string variable value is true or no?
func isTrueString(str string) bool {
	if str == "" || str == "0" {
		return true
	}
	return false
}
