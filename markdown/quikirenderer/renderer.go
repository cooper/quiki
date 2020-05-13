package quikirenderer

// Modeled from the html renderer at
// https://github.com/yuin/goldmark/blob/master/renderer/html/html.go
//
// Copyright (c) 2020 Mitchell Cooper
// Copyright (c) 2019 Yusuke Inuzuka
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// A Config struct has configurations for the quiki markup renderer.
type Config struct {
	Writer          Writer
	HardWraps       bool
	Unsafe          bool
	PartialPage     bool
	TableOfContents bool
	PageTitle       string
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		Writer:          DefaultWriter,
		HardWraps:       false,
		Unsafe:          false,
		PartialPage:     false,
		TableOfContents: false,
		PageTitle:       "",
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optHardWraps:
		c.HardWraps = value.(bool)
	case optUnsafe:
		c.Unsafe = value.(bool)
	case optPartialPage:
		c.PartialPage = value.(bool)
	case optTableOfContents:
		c.TableOfContents = value.(bool)
	case optPageTitle:
		c.PageTitle = value.(string)
	case optTextWriter:
		c.Writer = value.(Writer)
	}
}

// An Option interface sets options for the quiki markup renderer.
type Option interface {
	SetQuikiOption(*Config)
}

// TextWriter is an option name used in WithWriter.
const optTextWriter renderer.OptionName = "Writer"

type withWriter struct {
	value Writer
}

func (o *withWriter) SetConfig(c *renderer.Config) {
	c.Options[optTextWriter] = o.value
}

func (o *withWriter) SetQuikiOption(c *Config) {
	c.Writer = o.value
}

// WithWriter is a functional option that allow you to set the given writer to
// the renderer.
func WithWriter(writer Writer) interface {
	renderer.Option
	Option
} {
	return &withWriter{writer}
}

// HardWraps is an option name used in WithHardWraps.
const optHardWraps renderer.OptionName = "HardWraps"

type withHardWraps struct {
}

func (o *withHardWraps) SetConfig(c *renderer.Config) {
	c.Options[optHardWraps] = true
}

func (o *withHardWraps) SetQuikiOption(c *Config) {
	c.HardWraps = true
}

// WithHardWraps is a functional option that indicates whether softline breaks
// should be rendered as '<br>'.
func WithHardWraps() interface {
	renderer.Option
	Option
} {
	return &withHardWraps{}
}

// Unsafe is an option name used in WithUnsafe.
const optUnsafe renderer.OptionName = "Unsafe"

type withUnsafe struct {
}

func (o *withUnsafe) SetConfig(c *renderer.Config) {
	c.Options[optUnsafe] = true
}

func (o *withUnsafe) SetQuikiOption(c *Config) {
	c.Unsafe = true
}

// WithUnsafe is a functional option that renders dangerous contents
// (raw htmls and potentially dangerous links) as it is.
func WithUnsafe() interface {
	renderer.Option
	Option
} {
	return &withUnsafe{}
}

// PartialPage is an option name used in WithPartialPage.
const optPartialPage renderer.OptionName = "PartialPage"

type withPartialPage struct {
}

func (o *withPartialPage) SetConfig(c *renderer.Config) {
	c.Options[optPartialPage] = true
}

func (o *withPartialPage) SetQuikiOption(c *Config) {
	c.PartialPage = true
}

// WithPartialPage is a functional option that renders the Markdown
// as a portion of a page, without including quiki `@page` variables.
func WithPartialPage() interface {
	renderer.Option
	Option
} {
	return &withPartialPage{}
}

// TableOfContents is an option name used in WithTableOfContents.
const optTableOfContents renderer.OptionName = "TableOfContents"

type withTableOfContents struct {
}

func (o *withTableOfContents) SetConfig(c *renderer.Config) {
	c.Options[optTableOfContents] = true
}

func (o *withTableOfContents) SetQuikiOption(c *Config) {
	c.TableOfContents = true
}

// WithTableOfContents is a functional option that renders a table
// of contents on the page.
func WithTableOfContents() interface {
	renderer.Option
	Option
} {
	return &withTableOfContents{}
}

// PageTitle is an option name used in WithPageTitle.
const optPageTitle renderer.OptionName = "PageTitle"

type withPageTitle struct {
	title string
}

func (o *withPageTitle) SetConfig(c *renderer.Config) {
	c.Options[optPageTitle] = o.title
}

func (o *withPageTitle) SetQuikiOption(c *Config) {
	c.PageTitle = o.title
}

// WithPageTitle is a functional option that renders the `@page.title`
// variable to the provided text.
func WithPageTitle(title string) interface {
	renderer.Option
	Option
} {
	return &withPageTitle{title: title}
}

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as quiki markup.
type Renderer struct {
	Config
	headingLevel int
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
	}

	for _, opt := range opts {
		opt.SetQuikiOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		r.Writer.RawWrite(w, line.Value(source))
	}
}

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// head of page

		// partial page means don't include vars
		if r.PartialPage {
			return ast.WalkContinue, nil
		}

		// all the vars minus the title (has to be postposed til it is extracted)
		io.WriteString(w, "@page.author:       Markdown;\n")
		io.WriteString(w, "@page.generator:    quiki/markdown/goldmark;\n")
		io.WriteString(w, "@page.generated;\n\n")

		// table of contents
		if r.TableOfContents {
			io.WriteString(w, "toc{}\n\n")
		}
	} else {
		// foot of page

		// close open sections
		for ; r.headingLevel > 0; r.headingLevel-- {
			io.WriteString(w, "\n}")
		}

		// page title
		if r.PageTitle != "" {
			io.WriteString(w, "\n@page.title: "+quikiEscFmt(r.PageTitle)+";\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {

		// if we already have a header of this level open, this
		// terminates it. if we have a header of a lower level (higher
		// number) open, this terminates it and all others up to the
		// biggest level.
		for i := n.Level; i <= r.headingLevel; i++ {
			io.WriteString(w, "\n}\n")
		}

		// e.g. going from # to ###
		if n.Level > r.headingLevel+1 {
			for i := r.headingLevel + 2; i <= n.Level; i++ {
				io.WriteString(w, "~sec {\n")
			}
		}

		// set level, start the section with name opening tag.
		r.headingLevel = n.Level
		io.WriteString(w, "~sec [")

	} else {
		io.WriteString(w, "]")

		// TODO: assume page title as first heading
		// if r.PageTitle == "" {
		// 	r.PageTitle = r.heading
		// }

		// TODO: figure the anchor for github compatibility
		// id := node.HeadingID
		// if node.HeadingID == "" {
		// 	// https://github.com/jch/html-pipeline/blob/master/lib/html/pipeline/toc_filter.rb
		// 	// $section_id =~ tr/A-Z/a-z/;                 # ASCII downcase
		// 	id = strings.ToLower(r.heading)                // downcase
		// 	id = punctuationRegex.ReplaceAllString(id, "") // remove punctuation
		// 	id = strings.Replace(id, " ", "-", -1)         // replace spaces with dashes
		// 	r.heading = ""
		// }

		// // heading ID
		// id = r.ensureUniqueHeadingID(id)
		// if r.HeadingIDPrefix != "" {
		// 	id = r.HeadingIDPrefix + id
		// }
		// if r.HeadingIDSuffix != "" {
		// 	id = id + r.HeadingIDSuffix
		// }
		// r.addText(w, " "+id+"# ")
		io.WriteString(w, " {\n")
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<blockquote>\n")
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<pre><code>")
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("<pre><code")
		language := n.Language(source)
		if language != nil {
			_, _ = w.WriteString(" class=\"language-")
			r.Writer.Write(w, language)
			_, _ = w.WriteString("\"")
		}
		_ = w.WriteByte('>')
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		if r.Unsafe {
			l := n.Lines().Len()
			for i := 0; i < l; i++ {
				line := n.Lines().At(i)
				_, _ = w.Write(line.Value(source))
			}
		} else {
			_, _ = w.WriteString("<!-- raw HTML omitted -->\n")
		}
	} else {
		if n.HasClosure() {
			if r.Unsafe {
				closure := n.ClosureLine
				_, _ = w.Write(closure.Value(source))
			} else {
				_, _ = w.WriteString("<!-- raw HTML omitted -->\n")
			}
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(tag)
		if n.IsOrdered() && n.Start != 1 {
			fmt.Fprintf(w, " start=\"%d\"", n.Start)
		}
		_, _ = w.WriteString(">\n")
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<li>")
		fc := n.FirstChild()
		if fc != nil {
			if _, ok := fc.(*ast.TextBlock); !ok {
				_ = w.WriteByte('\n')
			}
		}
	} else {
		_, _ = w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<p>")
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("<hr")
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<a href="`)
	url := n.URL(source)
	label := n.Label(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(url), []byte("mailto:")) {
		_, _ = w.WriteString("mailto:")
	}
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(url, false)))
	_, _ = w.WriteString(`">`)
	_, _ = w.Write(util.EscapeHTML(label))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<code>")
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				r.Writer.RawWrite(w, value[:len(value)-1])
				if c != n.LastChild() {
					r.Writer.RawWrite(w, []byte(" "))
				}
			} else {
				r.Writer.RawWrite(w, value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("</code>")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	tag := "em"
	if n.Level == 2 {
		tag = "strong"
	}
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(tag)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag)
		_ = w.WriteByte('>')
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		if r.Unsafe || !IsDangerousURL(n.Destination) {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")
	if r.Unsafe || !IsDangerousURL(n.Destination) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(n.Text(source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString(">")
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	if r.Unsafe {
		n := node.(*ast.RawHTML)
		l := n.Segments.Len()
		for i := 0; i < l; i++ {
			segment := n.Segments.At(i)
			_, _ = w.Write(segment.Value(source))
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("<!-- raw HTML omitted -->")
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	if n.IsRaw() {
		r.Writer.RawWrite(w, segment.Value(source))
	} else {
		r.Writer.Write(w, segment.Value(source))
		if n.HardLineBreak() || (n.SoftLineBreak() && r.HardWraps) {
			_, _ = w.WriteString("<br>\n")
		} else if n.SoftLineBreak() {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	if n.IsCode() {
		_, _ = w.Write(n.Value)
	} else {
		if n.IsRaw() {
			r.Writer.RawWrite(w, n.Value)
		} else {
			r.Writer.Write(w, n.Value)
		}
	}
	return ast.WalkContinue, nil
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
			_, _ = writer.Write(v)
			return
		}
	}
	_, _ = writer.WriteRune(util.ToValidRune(r))
}

func (d *defaultWriter) RawWrite(writer util.BufWriter, source []byte) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		v := util.EscapeHTMLByte(source[i])
		if v != nil {
			_, _ = writer.Write(source[i-n : i])
			n = 0
			_, _ = writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		_, _ = writer.Write(source[l-n:])
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
