// Copyright (c) 2017, Mitchell Cooper
package config

import (
	"bytes"
	"log"
)

// buffer types
type buffType uint8

const (
	NO_BUF     buffType = iota // no buffer
	VAR_NAME                   // variable name
	VAR_VALUE                  // string variable value
	VAR_FORMAT                 // formatted text in between square brackets
)

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
	line         uint         // line number
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
	buf := bufferInfo{new(bytes.Buffer), t}
	state.buffers = append(state.buffers, buf)
}

// destroy a buffer, returning its contents
func (state *parserState) endBuffer() string {

	// there are no buffers; this is a bug
	if len(state.buffers) == 0 {
		log.Fatal("config: endBuffer() called with no buffers")
	}

	// pop the last buffer
	bufs := state.buffers
	buf, bufs := bufs[len(bufs)-1], bufs[:len(bufs)-1]
	state.buffers = bufs

	return buf.buffer.String()
}

// destroy the current variable state, returning the variable name
func (state *parserState) getVariable() string {
	name := state.varName
	state.varName = ""
	state.varPercent = false
	return name
}
