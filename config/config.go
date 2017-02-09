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
	Path string
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

// destroy a buffer
func (state *parserState) endBuffer() string {
	str := state.buffer.String()
	state.buffer = nil
	state.buffType = NO_BUF
	return str
}

// new config
func New(path string) *Config {
	return &Config{
        Path: path,
        vars: make(map[string]interface{}),
    }
}


// parse config
func (conf *Config) Parse() error {

	// open the config
	file, err := os.Open(conf.Path)
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
		log.Println("Increasing comment level")
		state.increaseCommentLevel()
		return nil

	} else if state.lastByte == '*' && b == '/' {

		// comment closure
		log.Println("Decreasing comment level")
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
	case '@':

		// we're already in a variable name
		if state.buffType == VAR_NAME {
			return errors.New("Invalid @ in variable name @" + state.buffer.String())
		}

		// we're not in a value, so this starts a variable
		if state.buffType != NO_BUF {
			goto realDefault
		}
		log.Println("Starting variable name")
		state.startBuffer(VAR_NAME)

	// end of variable name
	case ':':

		// we're in the var name, so terminate it
		if state.buffType != VAR_NAME {
			goto realDefault
		}
		name := state.endBuffer()
		log.Println("Got variable name: @" + name)
		state.startBuffer(VAR_VALUE)

	// end of a variable definition
	case ';':

		if state.buffType == VAR_NAME {

			// terminating a boolean
			state.endBuffer()
			log.Println("Got boolean")

		} else if state.buffType == VAR_VALUE {

			// terminating a string
			str := strings.TrimSpace(state.endBuffer())
			log.Println("Got string: " + str)

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

// return the map
func getWhere(where map[string]interface{}, name string) map[string]interface{} {

    // find interface
    iface := where[name]
    if iface == nil {
        return nil
    }

    // find map
    switch aMap := iface.(type) {
    case map[string]interface{}:
        return aMap
    }

    return nil
}

// get string value
func (conf *Config) Get(varName string) string {

    // split up into parts
    var parts = strings.Split(varName, ".")
    if len(parts) == 0 {
        return ""
    }

    // the last one is the final variable name
    lastPart, parts := parts[len(parts)-1], parts[:len(parts)-1]

    // start with the main map
    where := conf.vars

    // for each part, fetch the map inside
    for _, part := range parts {
        where = getWhere(where, part)
        log.Println("PART: " + part, " -> ", where)

        // nothing there, give up
        if where == nil {
            return ""
        }
    }

    // get the string value
    iface := where[lastPart]
    if iface == nil {
        return ""
    }
	return iface.(string)
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
