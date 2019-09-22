package wikifier

import "regexp"

type variableName struct {
	parent catch
	*genericCatch
}

func newVariableName(pfx string, pos position) *variableName {
	pc := []positionedContent{{pfx, pos}}
	return &variableName{genericCatch: &genericCatch{positionedPrefixContent: pc}}
}

func (vn *variableName) catchType() string {
	return catchTypeVariableName
}

func (vn *variableName) getParentCatch() catch {
	return vn.parent
}

// word-like chars and periods are OK in var names
func (vn *variableName) byteOK(b byte) bool {
	ok, _ := regexp.Match(`[\w\.]`, []byte{b})
	return ok
}

// skip whitespace in variable name
func (vn *variableName) shouldSkipByte(b byte) bool {
	skip, _ := regexp.Match(`\s`, []byte{b})
	return skip
}

type variableValue struct {
	parent catch
	*genericCatch
}

func newVariableValue() *variableValue {
	return &variableValue{genericCatch: &genericCatch{}}
}

func (vv *variableValue) catchType() string {
	return catchTypeVariableValue
}

func (vv *variableValue) getParentCatch() catch {
	return vv.parent
}

// word-like chars and periods are OK in var names
func (vv *variableValue) byteOK(b byte) bool {
	ok, _ := regexp.Match(`.`, []byte{b})
	return ok
}

// skip whitespace in variable name
func (vv *variableValue) shouldSkipByte(b byte) bool {
	return false
}
