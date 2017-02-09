package config

import (
    "os"
    "io"
    "log"
)

type Config struct {
    Path    string
}

type parserState struct {
    lastByte        byte
    commentLevel    uint8
    inComment       bool
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
    log.Println(string(b))
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
