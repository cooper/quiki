// Copyright (c) 2017, Mitchell Cooper
package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// parse config
func (conf *Config) Parse() error {

	// open the config
	file, err := os.Open(conf.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// initial state
	state := &parserState{
		line:    1,
		buffers: make([]bufferInfo, 0, 3),
	}
	conf.line = &state.line

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

		// byte error
		if err != nil {
			err = errors.New(fmt.Sprintf("%s:%d: %s", conf.path, *conf.line, err.Error()))
			return err
		}
	}

	state.line = 0
	return nil
}

// produce a warning
func (conf *Config) Warn(msg string) {
	log.Printf(conf.getWarn(msg))
}

func (conf *Config) getWarn(msg string) (res string) {
	line := *conf.line
	if line == 0 {
		res = fmt.Sprintf("%s: %s\n", conf.path, msg)
		return
	}
	res = fmt.Sprintf("%s:%d: %s\n", conf.path, line, msg)
	return
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

		// start of a text format
	case '[':

		// we're already in a text format.
		// this is supported by wikifier but not here
		if state.buffType() == VAR_FORMAT {
			return errors.New("Square brackets in format not yet supported")
		}

		// we aren't in a variable value, or maybe
		// the current variable does not allow interpolation
		if state.buffType() != VAR_VALUE || state.varPercent {
			goto realDefault
		}

		// otherwise, this starts a formatting token
		state.startBuffer(VAR_FORMAT)

	// end of a text format
	case ']':

		// not in a format
		if state.buffType() != VAR_FORMAT {
			goto realDefault
		}

		// otherwise, this terminates a formatting token
		tok := state.endBuffer()

		// parse the formatting token
		err, newVal := conf.getFormattingToken(tok, false)
		if err != nil {
			return err
		}
		if newVal == "" {
			conf.Warn("[" + tok + "] yields empty string")
		}

		// add the value returned by it to the variable value buffer
		state.buffer().WriteString(newVal)

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

	case '\n':
		state.line++
		goto realDefault

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

// return the value of a formatting token
func (conf *Config) getFormattingToken(tok string, disableVars bool) (error, string) {

	// normal variable
	if strings.HasPrefix(tok, "@") {
		if disableVars {
			goto badVariable
		}
		return nil, conf.Get(strings.TrimPrefix(tok, "@"))
	}

	// interpolable variable
	if strings.HasPrefix(tok, "%") {
		if disableVars {
			goto badVariable
		}
		val := strings.TrimPrefix("tok", "%")
		return conf.getFormattingToken(val, true)
	}

	return errors.New("Unknown formatting token [" + tok + "]"), ""

badVariable:
	return errors.New("Recursive variable " + tok + " detected"), ""
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
