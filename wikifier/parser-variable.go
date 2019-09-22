package wikifier

import "regexp"

type parserVariableName struct {
	parent parserCatch
	*genericCatch
}

func newParserVariableName(pfx string, pos parserPosition) *parserVariableName {
	pc := []positionedContent{{pfx, pos}}
	return &parserVariableName{genericCatch: &genericCatch{positionedPrefixContent: pc}}
}

func (vn *parserVariableName) catchType() string {
	return catchTypeVariableName
}

func (vn *parserVariableName) getParentCatch() parserCatch {
	return vn.parent
}

// word-like chars and periods are OK in var names
func (vn *parserVariableName) byteOK(b byte) bool {
	ok, _ := regexp.Match(`[\w\.]`, []byte{b})
	return ok
}

// skip whitespace in variable name
func (vn *parserVariableName) shouldSkipByte(b byte) bool {
	skip, _ := regexp.Match(`\s`, []byte{b})
	return skip
}

type parserVariableValue struct {
	parent parserCatch
	*genericCatch
}

func newParserVariableValue() *parserVariableValue {
	return &parserVariableValue{genericCatch: &genericCatch{}}
}

func (vv *parserVariableValue) catchType() string {
	return catchTypeVariableValue
}

func (vv *parserVariableValue) getParentCatch() parserCatch {
	return vv.parent
}

// word-like chars and periods are OK in var names
func (vv *parserVariableValue) byteOK(b byte) bool {
	ok, _ := regexp.Match(`.`, []byte{b})
	return ok
}

// skip whitespace in variable name
func (vv *parserVariableValue) shouldSkipByte(b byte) bool {
	return false
}
