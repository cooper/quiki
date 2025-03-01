// Package markdown translates Markdown to quiki source code.
package markdown

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/cooper/quiki/adminifier/utils"
	"github.com/russross/blackfriday/v2"
)

var punctuationRegex = regexp.MustCompile(`[^\w\- ]`)

// Run parses Markdown and renders quiki soure code.
func Run(input []byte) []byte {
	r := NewQuikiRenderer(QuikiRendererParameters{Flags: TableOfContents})
	return blackfriday.Run(input, blackfriday.WithRenderer(r), blackfriday.WithExtensions(blackfriday.NoEmptyLineBeforeBlock|blackfriday.CommonExtensions))
}

// QuikiFlags is renderer configuration options.
type QuikiFlags int

// QuikiFlags configuration options.
const (
	QuikiFlagsNone      QuikiFlags = 0         // No flags
	SkipHTML            QuikiFlags = 1 << iota // Skip preformatted HTML blocks
	SkipImages                                 // Skip embedded images
	SkipLinks                                  // Skip all links
	PartialPage                                // If true, no @page vars at start
	TableOfContents                            // If true, include TOC
	FootnoteReturnLinks                        // Generate a link at the end of a footnote to return to the source
)

// QuikiRendererParameters allows you to tweak the behavior of a QuikiRenderer.
type QuikiRendererParameters struct {

	// path to prepend to relative URLs
	AbsolutePrefix string

	// add this text to each footnote anchor, to ensure uniqueness.
	FootnoteAnchorPrefix string

	// Show this text inside the <a> tag for a footnote return link, if the
	// HTML_FOOTNOTE_RETURN_LINKS flag is enabled. If blank, the string
	// <sup>[return]</sup> is used.
	//
	FootnoteReturnLinkContents string

	// If set, add this text to the front of each Heading ID, to ensure
	// uniqueness.
	HeadingIDPrefix string

	// If set, add this text to the back of each Heading ID, to ensure uniqueness.
	HeadingIDSuffix string
	// Increase heading levels: if the offset is 1, <h1> becomes <h2> etc.
	// Negative offset is also valid.
	// Resulting levels are clipped between 1 and 6.
	HeadingLevelOffset int

	// page title. defaults to the first heading in the document
	Title string

	// flags to customize the renderer's behavior
	Flags QuikiFlags
}

// QuikiRenderer is a type that implements the Renderer interface for quiki source code output.
//
// Do not create this directly, instead use the NewQuikiRenderer function.
type QuikiRenderer struct {
	QuikiRendererParameters

	// Track heading IDs to prevent ID collision in a single generation.
	headingIDs map[string]int

	heading     string // heading text stored until end of heading
	headerLevel int    // section depth
	indent      int    // indent level
	linkDest    string // link destination stored until end of link text

	lastOutputLen int
}

// NewQuikiRenderer creates and configures a QuikiRenderer object, which
// satisfies the Renderer interface.
func NewQuikiRenderer(params QuikiRendererParameters) *QuikiRenderer {

	if params.FootnoteReturnLinkContents == "" {
		params.FootnoteReturnLinkContents = `<sup>[return]</sup>`
	}

	return &QuikiRenderer{
		QuikiRendererParameters: params,
		headingIDs:              make(map[string]int),
	}
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

func (r *QuikiRenderer) ensureUniqueHeadingID(id string) string {
	for count, found := r.headingIDs[id]; found; count, found = r.headingIDs[id] {
		tmp := fmt.Sprintf("%s-%d", id, count+1)

		if _, tmpFound := r.headingIDs[tmp]; !tmpFound {
			r.headingIDs[id] = count + 1
			id = tmp
		} else {
			id = id + "-1"
		}
	}

	if _, found := r.headingIDs[id]; !found {
		r.headingIDs[id] = 0
	}

	return id
}

func (r *QuikiRenderer) addAbsPrefix(link []byte) []byte {
	if r.AbsolutePrefix != "" && isRelativeLink(link) && link[0] != '.' {
		newDest := r.AbsolutePrefix
		if link[0] != '/' {
			newDest += "/"
		}
		newDest += string(link)
		return []byte(newDest)
	}
	return link
}

func codeLanguage(info []byte) string {
	endOfLang := bytes.IndexAny(info, "\t ")
	if endOfLang < 0 {
		endOfLang = len(info)
	}
	return string(info[:endOfLang])
}

func (r *QuikiRenderer) tag(w io.Writer, name []byte, attrs []string) {
	w.Write(name)
	if len(attrs) > 0 {
		w.Write(spaceBytes)
		w.Write([]byte(strings.Join(attrs, " ")))
	}
	w.Write(gtBytes)
	r.lastOutputLen = 1
}

func footnoteRef(prefix string, node *blackfriday.Node) []byte {
	urlFrag := prefix + string(slugify(node.Destination))
	anchor := fmt.Sprintf(`<a href="#fn:%s">%d</a>`, urlFrag, node.NoteID)
	return []byte(fmt.Sprintf(`<sup class="footnote-ref" id="fnref:%s">%s</sup>`, urlFrag, anchor))
}

func footnoteItem(prefix string, slug []byte) []byte {
	return []byte(fmt.Sprintf(`<li id="fn:%s%s">`, prefix, slug))
}

func footnoteReturnLink(prefix, returnLink string, slug []byte) []byte {
	const format = ` <a class="footnote-return" href="#fnref:%s%s">%s</a>`
	return []byte(fmt.Sprintf(format, prefix, slug, returnLink))
}

func skipParagraphTags(node *blackfriday.Node) bool {
	grandparent := node.Parent.Parent
	if grandparent == nil || grandparent.Type != blackfriday.List {
		return false
	}
	tightOrTerm := grandparent.Tight || node.Parent.ListFlags&blackfriday.ListTypeTerm != 0
	return grandparent.Type == blackfriday.List && tightOrTerm
}

func cellAlignment(align blackfriday.CellAlignFlags) string {
	switch align {
	case blackfriday.TableAlignmentLeft:
		return "left"
	case blackfriday.TableAlignmentRight:
		return "right"
	case blackfriday.TableAlignmentCenter:
		return "center"
	default:
		return ""
	}
}

func (r *QuikiRenderer) out(w io.Writer, text []byte) {
	w.Write(text)
	r.lastOutputLen = len(text)
}

func (r *QuikiRenderer) addText(w io.Writer, text string) {
	// if r.indent > 0 {
	// 	indentStr := "    "
	// 	newText := ""
	// 	for _, line := range strings.Split(text, "\n") {
	// 		newText += line + "\n" + indentStr
	// 	}
	// 	text = newText
	// }
	r.out(w, []byte(text))
}

func (r *QuikiRenderer) cr(w io.Writer) {
	if r.lastOutputLen > 0 {
		r.out(w, nlBytes)
	}
}

var (
	nlBytes    = []byte{'\n'}
	gtBytes    = []byte{'>'}
	spaceBytes = []byte{' '}
)

var (
	hrTag         = []byte("<hr />")
	tableTag      = []byte(`<table class="q-table">`)
	tableCloseTag = []byte("</table>")
	tdTag         = []byte("<td")
	tdCloseTag    = []byte("</td>")
	thTag         = []byte("<th")
	thCloseTag    = []byte("</th>")
	theadTag      = []byte("<thead>")
	theadCloseTag = []byte("</thead>")
	tbodyTag      = []byte("<tbody>")
	tbodyCloseTag = []byte("</tbody>")
	trTag         = []byte("<tr>")
	trCloseTag    = []byte("</tr>")

	footnotesDivBytes      = []byte("\n<div class=\"footnotes\">\n\n")
	footnotesCloseDivBytes = []byte("\n</div>\n")
)

// RenderNode is a default renderer of a single node of a syntax tree. For
// block nodes it will be called twice: first time with entering=true, second
// time with entering=false, so that it could know when it's working on an open
// tag and when on close. It writes the result to w.
//
// The return value is a way to tell the calling walker to adjust its walk
// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
// can ask the walker to skip a subtree of this node by returning SkipChildren.
// The typical behavior is to return GoToNext, which asks for the usual
// traversal to the next node.
func (r *QuikiRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	attrs := []string{}
	switch node.Type {

	case blackfriday.Text:
		s := string(node.Literal)
		if node.Parent.Type == blackfriday.Link {
			r.addText(w, utils.EscLink(s))
		} else if node.Parent.Type == blackfriday.Paragraph && node.Parent.Parent.Type == blackfriday.Item {
			r.addText(w, utils.EscListMapValue(s))
		} else if node.Parent.Type == blackfriday.Item {
			r.addText(w, utils.EscListMapValue(s))
		} else if node.Parent.Type == blackfriday.Heading {
			r.heading += s
			r.addText(w, utils.EscFmt(s))
		} else {
			r.addText(w, utils.EscFmt(s))
		}

	// newline
	case blackfriday.Softbreak:
		r.cr(w)
		// TODO: make it configurable via out(renderer.softbreak)

	// actual newline
	case blackfriday.Hardbreak:
		r.addText(w, "[br]")
		r.cr(w)

	// italicd
	case blackfriday.Emph:
		if entering {
			r.addText(w, "[i]")
		} else {
			r.addText(w, "[/i]")
		}

	// bold
	case blackfriday.Strong:
		if entering {
			r.addText(w, "[b]")
		} else {
			r.addText(w, "[/b]")
		}

	// strikethrough
	case blackfriday.Del:
		if entering {
			r.addText(w, "[s]")
		} else {
			r.addText(w, "[/s]")
		}

	// inline html
	case blackfriday.HTMLSpan:
		if r.Flags&SkipHTML != 0 {
			break
		}
		html := utils.EscFmt(string(node.Literal))
		r.addText(w, "[html:"+html+"]")

	// link
	case blackfriday.Link:
		// mark it but don't link it if it is not a safe link
		dest := node.LinkData.Destination
		if r.Flags&SkipLinks != 0 {
			if entering {
				r.addText(w, "[c]")
			} else {
				r.addText(w, "[/c]")
			}
		} else {
			if entering {
				link := string(r.addAbsPrefix(dest))
				link = utils.EscLink(link)
				if hashIdx := strings.IndexByte(link, '#'); hashIdx != -1 {
					r.linkDest = strings.TrimSuffix(link[:hashIdx], ".md") + link[hashIdx:]
				} else {
					r.linkDest = strings.TrimSuffix(link, ".md")
				}
				r.addText(w, "[[ ")

				// TODO: anything we can do with node.LinkData.Title?
			} else {
				// if node.NoteID != 0 {
				// 	break
				// }
				r.addText(w, " | "+r.linkDest+" ]]")
				r.linkDest = ""
			}
		}

	// image
	case blackfriday.Image:

		// configured to skip images
		if r.Flags&SkipImages != 0 {
			return blackfriday.SkipChildren
		}

		if entering {
			dest := node.LinkData.Destination
			dest = r.addAbsPrefix(dest)
			// FIXME: if dest is not relative, we can't display this image
			r.addText(w, "~image {\n    file: "+utils.Esc(string(dest))+";\n    alt: ")
		} else {
			// FIXME: can we do anything with node.LinkData.Title?
			r.out(w, []byte(";\n}"))
		}

	// inline code
	case blackfriday.Code:
		var code string
		if node.Parent != nil && node.Parent.Parent != nil &&
			node.Parent.Type == blackfriday.Paragraph && node.Parent.Parent.Type == blackfriday.Item {
			code = utils.EscListMapValue(string(node.Literal))
		} else {
			code = utils.EscFmt(string(node.Literal))
		}
		r.addText(w, "[c]"+code+"[/c]")

	// document
	case blackfriday.Document:
		// close any open sections
		if !entering {
			for ; r.headerLevel > 0; r.headerLevel-- {
				r.addText(w, "\n}")
			}
		}

	// paragraph
	case blackfriday.Paragraph:
		if skipParagraphTags(node) {
			break
		}
		if entering {
			// TODO: untangle this clusterfuck about when the newlines need
			// to be added and when not.
			if node.Prev != nil {
				switch node.Prev.Type {
				case blackfriday.HTMLBlock, blackfriday.List, blackfriday.Paragraph, blackfriday.Heading,
					blackfriday.CodeBlock, blackfriday.BlockQuote, blackfriday.HorizontalRule:
					r.cr(w)
				}
			}
			if node.Parent.Type == blackfriday.BlockQuote && node.Prev == nil {
				r.cr(w)
			}
			r.addText(w, "~p {\n")
			r.indent++
		} else {
			r.indent--
			r.addText(w, "\n}")
			if !(node.Parent.Type == blackfriday.Item && node.Next == nil) {
				r.cr(w)
			}
		}

	// blockquote
	case blackfriday.BlockQuote:
		if entering {
			r.cr(w)
			r.addText(w, "~quote {\n")
			r.indent++
		} else {
			r.indent--
			r.addText(w, "\n}")
			r.cr(w)
		}

	// HTML block
	case blackfriday.HTMLBlock:
		if r.Flags&SkipHTML != 0 {
			break
		}
		r.cr(w)
		r.addText(w, "~html {\n")
		r.indent++
		r.addText(w, utils.Esc(string(node.Literal)))
		r.indent--
		r.addText(w, "\n}")
		r.cr(w)

	// heading
	case blackfriday.Heading:
		level := r.QuikiRendererParameters.HeadingLevelOffset + node.Level
		if entering {

			// if we already have a header of this level open, this
			// terminates it. if we have a header of a lower level (higher
			// number) open, this terminates it and all others up to the
			// biggest level.
			for i := level; i <= r.headerLevel; i++ {
				r.indent--
				r.addText(w, "\n}\n")
			}

			// e.g. going from # to ###
			if level > r.headerLevel+1 {
				for i := r.headerLevel + 2; i <= level; i++ {
					r.indent++
					r.addText(w, "~sec {\n")
				}
			}

			// set level, start the section with name opening tag.
			r.headerLevel = level
			r.addText(w, "~sec [")

			// if node.IsTitleblock {
			// 	attrs = append(attrs, `class="title"`)
			// }

		} else {
			// r.out(w, closeTag)
			// if !(node.Parent.Type == blackfriday.Item && node.Next == nil) {
			// 	r.cr(w)
			// }

			// $indent++;
			r.indent++
			//     $add_text->("$current_text] {\n");
			r.addText(w, "]")

			// assume page title as first heading
			if r.Title == "" {
				r.Title = r.heading
			}

			// figure the anchor for github compatibility
			id := node.HeadingID
			if node.HeadingID == "" {
				// https://github.com/jch/html-pipeline/blob/master/lib/html/pipeline/toc_filter.rb
				// $section_id =~ tr/A-Z/a-z/;                 # ASCII downcase
				id = strings.ToLower(r.heading)                // downcase
				id = punctuationRegex.ReplaceAllString(id, "") // remove punctuation
				id = strings.Replace(id, " ", "-", -1)         // replace spaces with dashes
				r.heading = ""
			}

			// heading ID
			id = r.ensureUniqueHeadingID(id)
			if r.HeadingIDPrefix != "" {
				id = r.HeadingIDPrefix + id
			}
			if r.HeadingIDSuffix != "" {
				id = id + r.HeadingIDSuffix
			}
			r.addText(w, " "+id+"# {\n")
		}

	// horizontal rule
	// TODO
	case blackfriday.HorizontalRule:
		r.cr(w)
		r.out(w, hrTag)
		r.cr(w)

	case blackfriday.List:

		if entering {
			if node.IsFootnotesList {
				r.out(w, footnotesDivBytes)
				r.out(w, hrTag)
				r.cr(w)
			}
			r.cr(w)
			if node.Parent.Type == blackfriday.Item && node.Parent.Parent.Tight {
				r.cr(w)
			}

			if node.ListFlags&blackfriday.ListTypeOrdered != 0 {
				r.addText(w, "numlist {")
			} else if node.ListFlags&blackfriday.ListTypeDefinition != 0 {
				r.addText(w, "definitions {")
			} else {
				r.addText(w, "list {")
			}

			r.indent++
		} else {
			r.indent--
			r.addText(w, "\n}")
			// if node.Parent.Type == blackfriday.Item && node.Next != nil {
			// 	r.cr(w)
			// }
			// if node.Parent.Type == blackfriday.Document || node.Parent.Type == blackfriday.BlockQuote {
			// 	r.cr(w)
			// }
			if node.IsFootnotesList {
				r.out(w, footnotesCloseDivBytes)
			}
		}
	case blackfriday.Item:
		if entering {
			r.cr(w)
		} else {
			// if node.ListData.RefLink != nil {
			// 	slug := slugify(node.ListData.RefLink)
			// 	if r.Flags&FootnoteReturnLinks != 0 {
			// 		r.out(w, footnoteReturnLink(r.FootnoteAnchorPrefix, r.FootnoteReturnLinkContents, slug))
			// 	}
			// }
			r.addText(w, ";")
		}

	// code block
	case blackfriday.CodeBlock:
		r.cr(w)

		// TODO: count opening and closing brackets.
		// if they match, use brace-escape rather than utils.Esc()
		r.addText(w, "~code ")
		if lang := codeLanguage(node.Info); lang != "" {
			r.addText(w, "["+lang+"] ")
		}
		r.addText(w, "{\n")
		r.indent++
		r.addText(w, utils.Esc(string(node.Literal)))
		r.indent--
		r.addText(w, "}")

		if node.Parent.Type != blackfriday.Item {
			r.cr(w)
		}

	// table
	// just wrap in html
	case blackfriday.Table:
		if entering {
			r.cr(w)
			r.addText(w, "~html {")
			r.out(w, tableTag)
		} else {
			r.out(w, tableCloseTag)
			r.addText(w, "}")
			r.cr(w)
		}

	// table cell
	case blackfriday.TableCell:
		openTag := tdTag
		closeTag := tdCloseTag
		if node.IsHeader {
			openTag = thTag
			closeTag = thCloseTag
		}
		if entering {
			align := cellAlignment(node.Align)
			if align != "" {
				attrs = append(attrs, fmt.Sprintf(`align="%s"`, align))
			}
			if node.Prev == nil {
				r.cr(w)
			}
			r.tag(w, openTag, attrs)
		} else {
			r.out(w, closeTag)
			r.cr(w)
		}

	// table head
	case blackfriday.TableHead:
		if entering {
			r.cr(w)
			r.out(w, theadTag)
		} else {
			r.out(w, theadCloseTag)
			r.cr(w)
		}

	// table body
	case blackfriday.TableBody:
		if entering {
			r.cr(w)
			r.out(w, tbodyTag)
			// XXX: this is to adhere to a rather silly test. Should fix test.
			if node.FirstChild == nil {
				r.cr(w)
			}
		} else {
			r.out(w, tbodyCloseTag)
			r.cr(w)
		}

	// table row
	case blackfriday.TableRow:
		if entering {
			r.cr(w)
			r.out(w, trTag)
		} else {
			r.out(w, trCloseTag)
			r.cr(w)
		}

	// unknown
	default:
		panic("Unknown node type " + node.Type.String())
	}

	return blackfriday.GoToNext
}

// RenderFooter renders the page footer.
func (r *QuikiRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// title must be done after the heading is extracted
	if r.Title != "" {
		io.WriteString(w, "\n@page.title: "+utils.EscFmt(r.Title)+";\n")
	}
}

// RenderHeader renders the page header, which includes @page variable definitions.
func (r *QuikiRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	if r.Flags&PartialPage != 0 {
		return
	}
	io.WriteString(w, "@page.author:    Markdown;\n")
	io.WriteString(w, "@page.generator: quiki/markdown;\n")
	io.WriteString(w, "@page.generated;\n\n")
	if r.Flags&TableOfContents != 0 {
		io.WriteString(w, "toc{}\n\n")
	}
}
