package wikifier

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type codeBlock struct {
	*parserBlock
}

type quikiPreWrapper bool

func (p quikiPreWrapper) Start(code bool, styleAttr string) string {
	return fmt.Sprintf(`<pre class="q-code chroma">`)
}

func (p quikiPreWrapper) End(code bool) string {
	return "</pre>"
}

func newCodeBlock(name string, b *parserBlock) block {
	return &codeBlock{parserBlock: b}
}

func (cb *codeBlock) html(page *Page, el element) {
	el.setTag("pre")
	el.setMeta("noTags", true)
	el.setMeta("noIndent", true)

	// get code text
	text := ""
	for _, piece := range cb.textContent() {
		text += piece
	}

	// if block name is provided, it's the language
	var lexer chroma.Lexer
	if cb.blockName() != "" {
		lexer = lexers.Get(cb.blockName())
	}

	// did not find it by language name-- try to guess
	if lexer == nil {
		lexer = lexers.Analyse(text)
	}

	// still no match; fall back to plain text
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// coalesce like adjacent tokens
	lexer = chroma.Coalesce(lexer)

	// get style from page var or config
	pageStyle, _ := page.getPageStr("code.style")
	style := styles.Get(pageStyle)
	if style == nil {
		style = styles.Get(page.Opt.Page.Code.Style)
	}
	if style == nil {
		style = styles.Fallback
	}

	// create HTML formatter with separate CSS
	var cssBuilder, htmlBuilder strings.Builder
	formatter := html.New(html.WithClasses(true), html.WithPreWrapper(quikiPreWrapper(true)))

	// HTML
	iterator, err := lexer.Tokenise(nil, text)
	err = formatter.Format(&htmlBuilder, style, iterator)
	if err != nil {
		cb.warn(cb.openPosition(), err.Error())
	} else {
		el.addHTML(HTML(htmlBuilder.String()))
	}

	// CSS
	err = formatter.WriteCSS(&cssBuilder, style)
	if err != nil {
		cb.warn(cb.openPosition(), err.Error())
	} else if !page.codeStyles {
		page.staticStyles = append(page.staticStyles, cssBuilder.String())
		page.codeStyles = true
	}
}
