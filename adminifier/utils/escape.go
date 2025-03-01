package utils

import "strings"

// Esc escapes a string for use in a wikifier template
func Esc(s string) string {

	// escape existing escapes
	s = strings.Replace(s, "\\", "\\\\", -1)

	// ecape curly brackets
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "}", "\\}", -1)

	// fix comments (see wikifier#62)
	s = strings.Replace(s, "/*", "\\/*", -1)

	return s
}

// EscFmt escapes a string for use in wikifier formatted text
func EscFmt(s string) string {
	s = Esc(s)
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	return s
}

// EscLink is like EscFmt except also escapes pipe for [[ links ]]
func EscLink(s string) string {
	s = EscFmt(s)
	return strings.Replace(s, "|", "\\|", -1)
}

// EscListMapValue is like EscFmt except also escapes semicolon
func EscListMapValue(s string) string {
	s = EscFmt(s)
	return strings.Replace(s, ";", "\\;", -1)
}

// EscMapKey is like EscFmt except also escapes colon and semicolon
func EscMapKey(s string) string {
	s = EscListMapValue(s)
	return strings.Replace(s, ":", "\\:", -1)
}
