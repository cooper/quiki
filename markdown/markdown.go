package markdown

import (
	"bytes"
	"log"

	"github.com/cooper/quiki/markdown/quikirenderer"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Run parses Markdown and renders quiki soure code.
func Run(input []byte) []byte {
	md := goldmark.New(
		goldmark.WithRenderer(renderer.NewRenderer(renderer.WithNodeRenderers(
			util.Prioritized(quikirenderer.NewRenderer(), 1000),
		))),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			quikirenderer.WithTableOfContents(),
			quikirenderer.WithPartialPage(),
			quikirenderer.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(input, &buf); err != nil {
		log.Println(err)
		// TODO: proper error handling here
	}
	return buf.Bytes()
}
