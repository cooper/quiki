package quikirenderer

// Modeled from the html renderer at
// https://github.com/yuin/goldmark/blob/master/renderer/html/html.go
//
// Copyright (c) 2020 Mitchell Cooper
// Copyright (c) 2019 Yusuke Inuzuka
//
// See LICENSE

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/util"
)

// IsDangerousURL returns true if the given url seems a potentially dangerous url,
// otherwise false.
func IsDangerousURL(url []byte) bool {
	if bytes.HasPrefix(url, bDataImage) && len(url) >= 11 {
		v := url[11:]
		if bytes.HasPrefix(v, bPng) || bytes.HasPrefix(v, bGif) ||
			bytes.HasPrefix(v, bJpeg) || bytes.HasPrefix(v, bWebp) {
			return false
		}
		return true
	}
	return bytes.HasPrefix(url, bJs) || bytes.HasPrefix(url, bVb) ||
		bytes.HasPrefix(url, bFile) || bytes.HasPrefix(url, bData)
}

func quikiEsc(s string) string {

	// escape existing escapes
	s = strings.Replace(s, "\\", "\\\\", -1)

	// ecape curly brackets
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "}", "\\}", -1)

	// fix comments (see wikifier#62)
	s = strings.Replace(s, "/*", "\\/*", -1)

	return s
}

// like quikiEsc except also escapes formatting tags
func quikiEscFmt(s string) string {
	s = quikiEsc(s)
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	return s
}

// like quikiEscFmt except also escapes pipe for [[ links ]]
func quikiEscLink(s string) string {
	s = quikiEscFmt(s)
	return strings.Replace(s, "|", "\\|", -1)
}

// like quikiEscFmt except also escapes semicolon
func quikiEscListMapValue(s string) string {
	s = quikiEscFmt(s)
	return strings.Replace(s, ";", "\\;", -1)
}

// like quikiEscFmt except also escapes colon and semicolon
func quikiEscMapKey(s string) string {
	s = quikiEscListMapValue(s)
	return strings.Replace(s, ":", "\\:", -1)
}

func isRelativeLink(link []byte) (yes bool) {

	// section
	if link[0] == '#' {
		return true
	}

	// link begin with '/' but not '//', the second maybe a protocol relative link
	if len(link) >= 2 && link[0] == '/' && link[1] != '/' {
		return true
	}

	// only the root '/'
	if len(link) == 1 && link[0] == '/' {
		return true
	}

	// current directory : begin with "./"
	if bytes.HasPrefix(link, []byte("./")) {
		return true
	}

	// parent directory : begin with "../"
	if bytes.HasPrefix(link, []byte("../")) {
		return true
	}

	return false
}

var dataPrefix = []byte("data-")

// A Writer interface writes textual contents to a writer.
type Writer interface {
	// Write writes the given source to writer with resolving references and unescaping
	// backslash escaped characters.
	Write(writer util.BufWriter, source []byte)

	// RawWrite writes the given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWrite(writer util.BufWriter, source []byte)
}

type defaultWriter struct {
}

func escapeRune(writer util.BufWriter, r rune) {
	if r < 256 {
		v := util.EscapeHTMLByte(byte(r))
		if v != nil {
			writer.Write(v)
			return
		}
	}
	writer.WriteRune(util.ToValidRune(r))
}

func (d *defaultWriter) RawWrite(writer util.BufWriter, source []byte) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		v := util.EscapeHTMLByte(source[i])
		if v != nil {
			writer.Write(source[i-n : i])
			n = 0
			writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		writer.Write(source[l-n:])
	}
}

func (d *defaultWriter) Write(writer util.BufWriter, source []byte) {
	escaped := false
	var ok bool
	limit := len(source)
	n := 0
	for i := 0; i < limit; i++ {
		c := source[i]
		if escaped {
			if util.IsPunct(c) {
				d.RawWrite(writer, source[n:i-1])
				n = i
				escaped = false
				continue
			}
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit && source[next] == '#' {
				nnext := next + 1
				if nnext < limit {
					nc := source[nnext]
					// code point like #x22;
					if nnext < limit && nc == 'x' || nc == 'X' {
						start := nnext + 1
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsHexDecimal)
						if ok && i < limit && source[i] == ';' {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 16, 32)
							d.RawWrite(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
						// code point like #1234;
					} else if nc >= '0' && nc <= '9' {
						start := nnext
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsNumeric)
						if ok && i < limit && i-start < 8 && source[i] == ';' {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 0, 32)
							d.RawWrite(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
					}
				}
			} else {
				start := next
				i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsAlphaNumeric)
				// entity reference
				if ok && i < limit && source[i] == ';' {
					name := util.BytesToReadOnlyString(source[start:i])
					entity, ok := util.LookUpHTML5EntityByName(name)
					if ok {
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						d.RawWrite(writer, entity.Characters)
						continue
					}
				}
			}
			i = next - 1
		}
		if c == '\\' {
			escaped = true
			continue
		}
		escaped = false
	}
	d.RawWrite(writer, source[n:])
}

// DefaultWriter is a default implementation of the Writer.
var DefaultWriter = &defaultWriter{}

var bDataImage = []byte("data:image/")
var bPng = []byte("png;")
var bGif = []byte("gif;")
var bJpeg = []byte("jpeg;")
var bWebp = []byte("webp;")
var bJs = []byte("javascript:")
var bVb = []byte("vbscript:")
var bFile = []byte("file:")
var bData = []byte("data:")
