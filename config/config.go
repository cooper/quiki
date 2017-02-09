package config

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
)

type Config struct {
	Path string
}

const (
	NO_BUF    = iota // no buffer
	VAR_NAME  = iota // variable name
	VAR_VALUE = iota // string variable value
)

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
	state.removeLastCharacter()
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

func (state *parserState) removeLastCharacter() {
	if state.buffer == nil || state.buffer.Len() < 1 {
		return
	}
	state.buffer.Truncate(state.buffer.Len() - 1)
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

func New(path string) *Config {
	return &Config{Path: path}
}

func (conf *Config) Parse() error {

	// open the config
	file, err := os.Open(conf.Path)
	defer file.Close()
	if err != nil {
		return err
	}

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
		err = conf.handleCharacter(state, b[0])
		state.lastByte = b[0]

		// character error
		if err != nil {
			return err
		}
	}

	return nil
}

func (conf *Config) handleCharacter(state *parserState, b byte) error {

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

		// terminating a boolean
		if state.buffType == VAR_NAME {
			state.endBuffer()
			log.Println("Got boolean")
		} else if state.buffType == VAR_VALUE {
			str := state.endBuffer()
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

func (conf *Config) Get(varName string) string {
	return ""
}

func (conf *Config) GetBool(varName string) bool {
	if str := conf.Get(varName); str == "" {
		return false
	}
	return true
}
