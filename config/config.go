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
type buffType uint8

const (
	NO_BUF    = iota // no buffer
	VAR_NAME  = iota // variable name
	VAR_VALUE = iota // string variable value
)

// configuration, fetch conf values with conf.Get()
type Config struct {
	path string
	vars map[string]interface{}
}

// defines a buffer and its type
type bufferInfo struct {
	buffer   *bytes.Buffer
	buffType buffType
}

// parser state info
type parserState struct {
	lastByte     byte         // previous byte
	commentLevel buffType     // comment level
	inComment    bool         // true if in a comment
	escaped      bool         // true if the current byte is escaped
	buffers      []bufferInfo // buffers
	varName      string       // current variable name
	varPercent   bool         // true if current variable is a %var
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
	if state.buffer() == nil || state.buffer().Len() < n {
		return
	}
	state.buffer().Truncate(state.buffer().Len() - n)
}

// get the current buffer
func (state *parserState) buffer() *bytes.Buffer {

	// there are no buffers
	if len(state.buffers) == 0 {
		return nil
	}

	return state.buffers[len(state.buffers)-1].buffer
}

// get the current buffer type
func (state *parserState) buffType() buffType {

	// there are no buffers
	if len(state.buffers) == 0 {
		return NO_BUF
	}

	return state.buffers[len(state.buffers)-1].buffType
}

// start a new buffer
func (state *parserState) startBuffer(t buffType) {
	buff := bufferInfo{new(bytes.Buffer), t}
	state.buffers = append(state.buffers, buff)
}

// destroy a buffer, returning its contents
func (state *parserState) endBuffer() string {

	// there are no buffers
	if len(state.buffers) == 0 {
		log.Fatal("config: endBuffer() called with no buffers")
	}

	// pop the last buffer
	buffs := state.buffers
	buff, buffs := buffs[len(buffs)-1], buffs[:len(buffs)-1]
	state.buffers = buffs

	return buff.buffer.String()
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

	state := &parserState{
		buffers: make([]bufferInfo, 3),
	}
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
		if state.buffType() == VAR_NAME {
			return errors.New("Already in variable name @" + state.endBuffer())
		}

		// this is only allowed at the top level
		if state.buffType() != NO_BUF {
			goto realDefault
		}

		// we're not in a value, so this starts a variable name
		state.varPercent = b == '%'
		state.startBuffer(VAR_NAME)

	// end of variable name, start of string value
	case ':':

		// not in a variable name
		if state.buffType() != VAR_NAME {
			goto realDefault
		}

		// we're in the var name, so terminate it
		state.varName = state.endBuffer()
		state.startBuffer(VAR_VALUE)

	case '[':

		// we aren't in a variable value, or maybe
		// the current variable does not allow interpolation
		if state.buffType() != VAR_VALUE || state.varPercent {
			goto realDefault
		}

	// end of a variable definition
	case ';':

		if state.buffType() == VAR_NAME {

			// terminating a boolean
			state.endBuffer()
			conf.Set(state.getVariable(), "1")

		} else if state.buffType() == VAR_VALUE {

			// terminating a string
			value := strings.TrimSpace(state.endBuffer())
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
	if state.buffer() != nil {
		state.buffer().WriteByte(b)
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
