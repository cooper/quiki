package wikifier

import (
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	_ "github.com/cooper/ferret-chroma" // for ferret language support
)

type codeBlock struct {
	*parserBlock
}

type quikiPreWrapper bool

func init() {
	styles.Fallback = styles.Get("monokailight")
}

func (p quikiPreWrapper) Start(code bool, styleAttr string) string {
	return `<pre class="q-code chroma">`
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

	// if block name or page.code.lang is provided, it's the language
	var lexer chroma.Lexer
	if cb.blockName() != "" {
		lexer = lexers.Get(cb.blockName())
		if lexer == nil {
			cb.warn(cb.openPosition(), "No such code{} language '"+cb.blockName()+"'")
		}
	}
	if lexer == nil && page.Opt.Page.Code.Lang != "" {
		lexer = lexers.Get(page.Opt.Page.Code.Lang)
		if lexer == nil {
			cb.warn(cb.openPosition(), "No such code{} language '"+page.Opt.Page.Code.Lang+"' (from config)")
		}
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
	style := styles.Fallback
	pageStyle, _ := page.getPageStr("code.style")
	if pageStyle != "" {
		style = styles.Get(pageStyle)
		if style == styles.Fallback {
			cb.warn(cb.openPosition(), "No such code{} style '"+pageStyle+"'")
		}
	}
	if style == styles.Fallback && page.Opt.Page.Code.Style != "" {
		style = styles.Get(page.Opt.Page.Code.Style)
		if style == styles.Fallback {
			cb.warn(cb.openPosition(), "No such code{} style '"+page.Opt.Page.Code.Style+"' (from config)")
		}
	}

	// create HTML formatter with separate CSS
	var cssBuilder, htmlBuilder strings.Builder
	formatter := html.New(html.WithClasses(true), html.WithPreWrapper(quikiPreWrapper(true)))

	// HTML
	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		cb.warn(cb.openPosition(), err.Error())
	} else {
		err := formatter.Format(&htmlBuilder, style, iterator)
		if err != nil {
			cb.warn(cb.openPosition(), err.Error())
		} else {
			el.addHTML(HTML(htmlBuilder.String()))
		}
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
