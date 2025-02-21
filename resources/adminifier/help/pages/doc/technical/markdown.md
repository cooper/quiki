# markdown
--
    import "github.com/cooper/quiki/markdown"

Package markdown translates Markdown to quiki source code.

## Usage

#### func  Run

```go
func Run(input []byte) []byte
```
Run parses Markdown and renders quiki soure code.

#### type QuikiFlags

```go
type QuikiFlags int
```

QuikiFlags is renderer configuration options.

```go
const (
	QuikiFlagsNone      QuikiFlags = 0         // No flags
	SkipHTML            QuikiFlags = 1 << iota // Skip preformatted HTML blocks
	SkipImages                                 // Skip embedded images
	SkipLinks                                  // Skip all links
	PartialPage                                // If true, no @page vars at start
	TableOfContents                            // If true, include TOC
	FootnoteReturnLinks                        // Generate a link at the end of a footnote to return to the source
)
```
QuikiFlags configuration options.

#### type QuikiRenderer

```go
type QuikiRenderer struct {
	QuikiRendererParameters
}
```

QuikiRenderer is a type that implements the Renderer interface for quiki source
code output.

Do not create this directly, instead use the NewQuikiRenderer function.

#### func  NewQuikiRenderer

```go
func NewQuikiRenderer(params QuikiRendererParameters) *QuikiRenderer
```
NewQuikiRenderer creates and configures a QuikiRenderer object, which satisfies
the Renderer interface.

#### func (*QuikiRenderer) RenderFooter

```go
func (r *QuikiRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node)
```
RenderFooter renders the page footer.

#### func (*QuikiRenderer) RenderHeader

```go
func (r *QuikiRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node)
```
RenderHeader renders the page header, which includes @page variable definitions.

#### func (*QuikiRenderer) RenderNode

```go
func (r *QuikiRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus
```
RenderNode is a default renderer of a single node of a syntax tree. For block
nodes it will be called twice: first time with entering=true, second time with
entering=false, so that it could know when it's working on an open tag and when
on close. It writes the result to w.

The return value is a way to tell the calling walker to adjust its walk pattern:
e.g. it can terminate the traversal by returning Terminate. Or it can ask the
walker to skip a subtree of this node by returning SkipChildren. The typical
behavior is to return GoToNext, which asks for the usual traversal to the next
node.

#### type QuikiRendererParameters

```go
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
```

QuikiRendererParameters allows you to tweak the behavior of a QuikiRenderer.
