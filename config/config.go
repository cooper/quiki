package config

import (
    "os"
    "io"
    //"log"
    "bytes"
)

type Config struct {
    Path    string
}

type parserState struct {
    lastByte            byte            // previous byte
    commentLevel        uint8           // comment level
    inComment           bool            // true if in a comment
    escaped             bool            // true if the current byte is escaped
    variableNameBuf     *bytes.Buffer   // variable name buffer
    variableValueBuf    *bytes.Buffer   // variable value buffer
}

// increase block comment level
func (state parserState) increaseCommentLevel() {
    state.commentLevel++
    state.inComment = true
}

// decrease block comment level
func (state parserState) decreaseCommentLevel() {
    if state.commentLevel <= 0 {
        return
    }
    state.commentLevel--
    state.inComment = state.commentLevel != 0
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

    state := parserState{}
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

        // character error
        if err != nil {
            return err
        }
    }

    return nil
}

func (conf *Config) handleCharacter(state parserState, b byte) error {

    // this character is escaped
    if state.escaped {
        b = 0
        state.escaped = false
    }

    switch b {

    // comment entrance
    case '*':
        if state.lastByte == '/' {
            state.increaseCommentLevel()
            break
        }
        fallthrough

    // comment closure
    case '/':
        if state.lastByte == '*' {
            state.decreaseCommentLevel()
            break
        }
        fallthrough

    // escape
    case '\\':
        state.escaped = true

    default:

        // we're in a comment; ignore this
        if state.inComment {
            break
        }
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
