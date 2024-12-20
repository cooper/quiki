package wikifier

import "regexp"

var (
	variableNameRgx = regexp.MustCompile(`[\w\.\/]`)
	spaceRgx        = regexp.MustCompile(`\s`)
	anyCharRgx      = regexp.MustCompile(`.`)
)

type variableName struct {
	parent catch
	*genericCatch
}

func newVariableName(pfx string, pos Position) *variableName {
	pc := []posContent{{pfx, pos}}
	return &variableName{genericCatch: &genericCatch{positionedPrefix: pc}}
}

func (vn *variableName) catchType() catchType {
	return catchTypeVariableName
}

func (vn *variableName) parentCatch() catch {
	return vn.parent
}

// word-like chars and periods are OK in var names
func (vn *variableName) runeOk(r rune) bool {
	return variableNameRgx.MatchString(string(r))
}

// skip whitespace in variable name
func (vn *variableName) shouldSkipRune(r rune) bool {
	return spaceRgx.MatchString(string(r))
}

type variableValue struct {
	parent catch
	*genericCatch
}

func newVariableValue() *variableValue {
	return &variableValue{genericCatch: &genericCatch{}}
}

func (vv *variableValue) catchType() catchType {
	return catchTypeVariableValue
}

func (vv *variableValue) parentCatch() catch {
	return vv.parent
}

func (vv *variableValue) runeOk(r rune) bool {
	if r == '\n' {
		return true
	}
	return anyCharRgx.MatchString(string(r))
}

func (vv *variableValue) shouldSkipRune(rune) bool {
	return false
}
