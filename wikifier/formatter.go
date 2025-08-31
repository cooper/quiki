package wikifier

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/enescakir/emoji"
)

var (
	variableRegex = regexp.MustCompile(`^([@%])([\w\.\/]+)$`)
	linkRegex     = regexp.MustCompile(`^((\w+)://|\$)\s*`)
	mailRegex     = regexp.MustCompile(`^[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,63}$`)
	colorRegex    = regexp.MustCompile(`(?i)^#[\da-f]+$`)
	wikiRegex     = regexp.MustCompile(`^(\w+):(.*)$`)
	oldLinkRegex  = regexp.MustCompile(`^([\!\$\~]+?)(.+)([\!\$\~]+?)$`)
	emojiRegex    = regexp.MustCompile(`:\w+:`)
)

var colors = map[string]string{
	"aliceblue":            "#f0f8ff",
	"antiquewhite":         "#faebd7",
	"aqua":                 "#00ffff",
	"aquamarine":           "#7fffd4",
	"azure":                "#f0ffff",
	"beige":                "#f5f5dc",
	"bisque":               "#ffe4c4",
	"black":                "#000000",
	"blanchedalmond":       "#ffebcd",
	"blue":                 "#0000ff",
	"blueviolet":           "#8a2be2",
	"brown":                "#a52a2a",
	"burlywood":            "#deb887",
	"cadetblue":            "#5f9ea0",
	"chartreuse":           "#7fff00",
	"chocolate":            "#d2691e",
	"coral":                "#ff7f50",
	"cornflowerblue":       "#6495ed",
	"cornsilk":             "#fff8dc",
	"crimson":              "#dc143c",
	"cyan":                 "#00ffff",
	"darkblue":             "#00008b",
	"darkcyan":             "#008b8b",
	"darkgoldenrod":        "#b8860b",
	"darkgray":             "#a9a9a9",
	"darkgreen":            "#006400",
	"darkkhaki":            "#bdb76b",
	"darkmagenta":          "#8b008b",
	"darkolivegreen":       "#556b2f",
	"darkorange":           "#ff8c00",
	"darkorchid":           "#9932cc",
	"darkred":              "#8b0000",
	"darksalmon":           "#e9967a",
	"darkseagreen":         "#8fbc8f",
	"darkslateblue":        "#483d8b",
	"darkslategray":        "#2f4f4f",
	"darkturquoise":        "#00ced1",
	"darkviolet":           "#9400d3",
	"deeppink":             "#ff1493",
	"deepskyblue":          "#00bfff",
	"dimgray":              "#696969",
	"dodgerblue":           "#1e90ff",
	"firebrick":            "#b22222",
	"floralwhite":          "#fffaf0",
	"forestgreen":          "#228b22",
	"fuchsia":              "#ff00ff",
	"gainsboro":            "#dcdcdc",
	"ghostwhite":           "#f8f8ff",
	"gold":                 "#ffd700",
	"goldenrod":            "#daa520",
	"gray":                 "#808080",
	"green":                "#008000",
	"greenyellow":          "#adff2f",
	"honeydew":             "#f0fff0",
	"hotpink":              "#ff69b4",
	"indianred":            "#cd5c5c",
	"indigo":               "#4b0082",
	"ivory":                "#fffff0",
	"khaki":                "#f0e68c",
	"lavender":             "#e6e6fa",
	"lavenderblush":        "#fff0f5",
	"lawngreen":            "#7cfc00",
	"lemonchiffon":         "#fffacd",
	"lightblue":            "#add8e6",
	"lightcoral":           "#f08080",
	"lightcyan":            "#e0ffff",
	"lightgoldenrodyellow": "#fafad2",
	"lightgray":            "#d3d3d3",
	"lightgreen":           "#90ee90",
	"lightpink":            "#ffb6c1",
	"lightsalmon":          "#ffa07a",
	"lightseagreen":        "#20b2aa",
	"lightskyblue":         "#87cefa",
	"lightslategray":       "#778899",
	"lightsteelblue":       "#b0c4de",
	"lightyellow":          "#ffffe0",
	"lime":                 "#00ff00",
	"limegreen":            "#32cd32",
	"linen":                "#faf0e6",
	"magenta":              "#ff00ff",
	"maroon":               "#800000",
	"mediumaquamarine":     "#66cdaa",
	"mediumblue":           "#0000cd",
	"mediumorchid":         "#ba55d3",
	"mediumpurple":         "#9370db",
	"mediumseagreen":       "#3cb371",
	"mediumslateblue":      "#7b68ee",
	"mediumspringgreen":    "#00fa9a",
	"mediumturquoise":      "#48d1cc",
	"mediumvioletred":      "#c71585",
	"midnightblue":         "#191970",
	"mintcream":            "#f5fffa",
	"mistyrose":            "#ffe4e1",
	"moccasin":             "#ffe4b5",
	"navajowhite":          "#ffdead",
	"navy":                 "#000080",
	"oldlace":              "#fdf5e6",
	"olive":                "#808000",
	"olivedrab":            "#6b8e23",
	"orange":               "#ffa500",
	"orangered":            "#ff4500",
	"orchid":               "#da70d6",
	"palegoldenrod":        "#eee8aa",
	"palegreen":            "#98fb98",
	"paleturquoise":        "#afeeee",
	"palevioletred":        "#db7093",
	"papayawhip":           "#ffefd5",
	"peachpuff":            "#ffdab9",
	"peru":                 "#cd853f",
	"pink":                 "#ffc0cb",
	"plum":                 "#dda0dd",
	"powderblue":           "#b0e0e6",
	"purple":               "#800080",
	"red":                  "#ff0000",
	"rosybrown":            "#bc8f8f",
	"royalblue":            "#4169e1",
	"saddlebrown":          "#8b4513",
	"salmon":               "#fa8072",
	"sandybrown":           "#f4a460",
	"seagreen":             "#2e8b57",
	"seashell":             "#fff5ee",
	"sienna":               "#a0522d",
	"silver":               "#c0c0c0",
	"skyblue":              "#87ceeb",
	"slateblue":            "#6a5acd",
	"slategray":            "#708090",
	"snow":                 "#fffafa",
	"springgreen":          "#00ff7f",
	"steelblue":            "#4682b4",
	"tan":                  "#d2b48c",
	"teal":                 "#008080",
	"thistle":              "#d8bfd8",
	"tomato":               "#ff6347",
	"turquoise":            "#40e0d0",
	"violet":               "#ee82ee",
	"wheat":                "#f5deb3",
	"white":                "#ffffff",
	"whitesmoke":           "#f5f5f5",
	"yellow":               "#ffff00",
	"yellowgreen":          "#9acd32",
}

var staticFormats = map[string]string{
	"i":  `<span style="font-style: italic;">`, // italic
	"/i": `</span>`,

	"b":  `<span style="font-weight: bold;">`, // bold
	"/b": `</span>`,

	"s":  `<span style="text-decoration: line-through;">`, // strike
	"/s": `</span>`,

	"c":  `<code>`, // inline code
	"/c": `</code>`,

	"q":  `<span style="font-style: italic;">"`, // inline quote
	"/q": `"</span>`,

	"^":  `<sup>`, // superscript
	"/^": `</sup>`,

	"v":  `<sub>`, // subscript
	"/v": `</sub>`,

	"/": `</span>`, // ends a color

	"nl": `<br />`, // line break
	"br": `<br />`, // (deprecated)

	"--":  `&ndash;`, // en dash
	"---": `&mdash;`, // em dash
}

// FmtOpt describes options for page.FmtOpts.
type FmtOpt struct {
	Pos         Position // position used for warnings (set internally)
	NoEntities  bool     // disables html entity conversion
	NoWarnings  bool     // silence warnings for undefined variables
	noVariables bool     // set internally to prevent recursive interpolation
}

type formatParser struct {
	runes  []rune  // characters we're parsing through
	pos    int     // current position in runes
	block  block   // block context for variables and warnings
	opts   *FmtOpt // formatting options
	result []any   // accumulated string and HTML pieces
}

// format generates HTML from a quiki-encoded formatted string.
func format(b block, text string, pos Position) HTML {
	return formatOpts(b, text, pos, FmtOpt{})
}

// formatOpts is like format except you can specify additional options with the FmtOpt argument.
func formatOpts(b block, text string, pos Position, o FmtOpt) HTML {
	// let's not waste any time here
	if text == "" {
		return ""
	}

	o.Pos = pos
	// find and copy the position
	if o.Pos.none() && b.page().parser != nil {
		o.Pos = b.page().parser.pos
	}

	p := &formatParser{
		runes: []rune(convertMarkdown(text)), // convert markdown first
		block: b,
		opts:  &o,
	}
	return p.parse()
}

func (p *formatParser) parse() HTML {
	for p.pos < len(p.runes) {
		if !p.tryFormat() {
			p.appendChar()
		}
	}
	return p.buildHTML()
}

func (p *formatParser) tryFormat() bool {
	// check if we're at a format marker and not escaped
	if p.current() != '[' || p.escaped() {
		return false
	}

	start := p.pos
	if depth := p.findFormatEnd(); depth > 0 {
		// extract the format type between the brackets
		formatType := string(p.runes[start+depth : p.pos-depth])
		p.appendHTML(parseFormatType(p.block, formatType, p.opts))
		return true
	}

	// no matching bracket found, reset position
	p.pos = start
	return false
}

func (p *formatParser) findFormatEnd() int {
	depth := 0 // how far [[in]] we are
	p.pos++

	for p.pos < len(p.runes) {
		// skip escaped characters entirely
		if p.escaped() {
			p.pos += 2
			continue
		}

		if p.current() == '[' {
			depth++ // going deeper into nested brackets
		} else if p.current() == ']' {
			if depth == 0 {
				p.pos++ // consume the closing bracket
				return 1
			}
			depth-- // coming back out
		}
		p.pos++
	}
	return 0 // no matching bracket found
}

func (p *formatParser) appendChar() {
	char := p.current()

	// check if this character should be escaped based on previous backslash
	wasEscaped := p.escaped()

	// after convertMarkdown, backslashes should be treated as literal characters
	// only skip backslashes if they're actually escaping brackets
	if char == '\\' && !wasEscaped {
		// check if next char is something that should be escaped
		if p.pos+1 < len(p.runes) && (p.runes[p.pos+1] == '[' || p.runes[p.pos+1] == ']') {
			p.pos++ // skip the escape character for brackets
			return
		}
		// otherwise, treat backslash as literal character (fall through)
	}

	// if the previous char was a backslash, this char is escaped, just include it literally
	if wasEscaped {
		// include the escaped character
		if len(p.result) == 0 || p.isHTML(len(p.result)-1) {
			p.result = append(p.result, string(char))
		} else {
			p.result[len(p.result)-1] = p.result[len(p.result)-1].(string) + string(char)
		}
		p.pos++
		return
	}

	// normal character, append to existing string or create new one
	if len(p.result) == 0 || p.isHTML(len(p.result)-1) {
		p.result = append(p.result, string(char))
	} else {
		p.result[len(p.result)-1] = p.result[len(p.result)-1].(string) + string(char)
	}
	p.updatePos()
}

func (p *formatParser) appendHTML(html HTML) {
	p.result = append(p.result, html)
}

func (p *formatParser) buildHTML() HTML {
	// join the parts together, converting entities as needed
	var final strings.Builder
	for _, item := range p.result {
		switch v := item.(type) {
		case string:
			if p.opts.NoEntities {
				final.WriteString(v)
			} else {
				final.WriteString(html.EscapeString(v))
			}
		case HTML:
			final.WriteString(string(v))
		}
	}
	return HTML(final.String())
}

func (p *formatParser) current() rune {
	if p.pos >= len(p.runes) {
		return 0
	}
	return p.runes[p.pos]
}

func (p *formatParser) escaped() bool {
	// after convertMarkdown, single backslashes are literal characters
	// only sequences that convertMarkdown didn't handle (like \[) are escapes
	// check if previous character is backslash, but since convertMarkdown
	// already processed \\* and \\, any remaining \\ should be literal
	if p.pos == 0 || p.runes[p.pos-1] != '\\' {
		return false
	}

	// if we have a single backslash, treat it as literal (convertMarkdown result)
	// only odd numbers of consecutive backslashes indicate escaping
	backslashCount := 0
	for i := p.pos - 1; i >= 0 && p.runes[i] == '\\'; i-- {
		backslashCount++
	}

	// after convertMarkdown processing, we should be more conservative
	// only treat as escaped if we have an odd number of backslashes
	// and the current char is something that wasn't handled by convertMarkdown
	if backslashCount%2 == 1 {
		// only escape brackets, since convertMarkdown handled asterisks
		return p.current() == '[' || p.current() == ']'
	}
	return false
}

func (p *formatParser) isHTML(idx int) bool {
	_, ok := p.result[idx].(HTML)
	return ok
}

func (p *formatParser) updatePos() {
	// track line and column for error reporting
	if p.current() == '\n' {
		p.opts.Pos.Line++
		p.opts.Pos.Column = 0
	} else {
		p.opts.Pos.Column++
	}
	p.pos++
}

// convertMarkdown processes markdown-style asterisk formatting before quiki parsing
// convertMarkdown processes markdown-style asterisk formatting before quiki parsing
func convertMarkdown(text string) string {
	runes := []rune(text)
	var result strings.Builder

	for i := 0; i < len(runes); {
		// handle escaped backslashes - \\ becomes literal \
		if i+1 < len(runes) && runes[i] == '\\' && runes[i+1] == '\\' {
			result.WriteRune('\\') // output literal backslash, consume both
			i += 2
			continue
		}

		// handle escaped asterisks - \* becomes literal *
		if i+1 < len(runes) && runes[i] == '\\' && runes[i+1] == '*' {
			result.WriteRune('*') // output literal asterisk, consume the backslash
			i += 2
			continue
		}

		// handle single backslash at end or before newline - becomes literal \
		if runes[i] == '\\' && (i+1 >= len(runes) || runes[i+1] == '\n') {
			result.WriteRune('\\') // output literal backslash
			i++
			continue
		}

		// handle literal asterisks that follow escaped ones
		// if we just processed an escape, treat next * as literal too
		if runes[i] == '*' && i >= 2 && runes[i-2] == '\\' && runes[i-1] == '*' {
			result.WriteRune('*')
			i++
			continue
		}

		// try bold first (longer match) - need at least 3 stars for ***
		if consumed := tryConvertBold(runes, i, &result); consumed > 0 {
			i += consumed
			continue
		}

		// then try italic
		if consumed := tryConvertItalic(runes, i, &result); consumed > 0 {
			i += consumed
			continue
		}

		result.WriteRune(runes[i])
		i++
	}

	output := result.String()
	return output
}

func tryConvertBold(runes []rune, pos int, result *strings.Builder) int {
	// look for **text** but handle ***text*** specially
	if runes[pos] != '*' || pos+1 >= len(runes) || runes[pos+1] != '*' {
		return 0
	}

	// check for *** (bold italic)
	if pos+2 < len(runes) && runes[pos+2] == '*' {
		if end := findMarkdownEnd(runes, pos+3, "***"); end != -1 {
			// convert ***text*** to [b][i]text[/i][/b]
			result.WriteString("[b][i]")
			result.WriteString(string(runes[pos+3 : end]))
			result.WriteString("[/i][/b]")
			return end + 3 - pos
		}
		return 0
	}

	// regular **text**
	if end := findMarkdownEnd(runes, pos+2, "**"); end != -1 {
		result.WriteString("[b]")
		result.WriteString(convertMarkdown(string(runes[pos+2 : end]))) // recursively handle nested
		result.WriteString("[/b]")
		return end + 2 - pos
	}
	return 0
}

func tryConvertItalic(runes []rune, pos int, result *strings.Builder) int {
	// look for *text* but not **text** or ***text***
	if runes[pos] != '*' {
		return 0
	}

	// don't match if this is part of ** or ***
	if pos+1 < len(runes) && runes[pos+1] == '*' {
		return 0
	}

	if end := findMarkdownEnd(runes, pos+1, "*"); end != -1 {
		result.WriteString("[i]")
		result.WriteString(string(runes[pos+1 : end])) // no recursion for italic
		result.WriteString("[/i]")
		return end + 1 - pos
	}
	return 0
}

func findMarkdownEnd(runes []rune, start int, delimiter string) int {
	delim := []rune(delimiter)

	for i := start; i <= len(runes)-len(delim); {
		// skip escaped asterisks - handle \\* sequences
		if i+1 < len(runes) && runes[i] == '\\' && runes[i+1] == '*' {
			i += 2 // skip the escaped asterisk
			continue
		}

		// markdown doesn't span lines
		if runes[i] == '\n' {
			return -1
		}

		// check if we found the delimiter
		match := true
		for j, r := range delim {
			if i+j >= len(runes) || runes[i+j] != r {
				match = false
				break
			}
		}
		if match {
			return i
		}

		i++
	}
	return -1
}

func parseFormatType(b block, formatType string, o *FmtOpt) HTML {
	// static format
	if format, exists := staticFormats[strings.ToLower(formatType)]; exists {
		return HTML(format)
	}

	// variable
	if !o.noVariables && variableRegex.MatchString(formatType) {
		return handleVariable(b, formatType, o)
	}

	// html entity
	if formatType[0] == '&' {
		return HTML("&" + formatType[1:] + ";")
	}

	// link
	if formatType[0] == '[' && formatType[len(formatType)-1] == ']' {
		return handleLink(b, formatType, o)
	}

	// color
	if color, exists := colors[strings.ToLower(formatType)]; exists {
		return HTML(`<span style="color: ` + color + `;">`)
	}
	if colorRegex.MatchString(formatType) {
		return HTML(`<span style="color: ` + formatType + `;">`)
	}

	// inline html
	if strings.HasPrefix(formatType, "html:") {
		return HTML(strings.TrimPrefix(formatType, "html:"))
	}

	// emoji
	if emojiRegex.MatchString(formatType) {
		if found, ok := emoji.Map()[formatType]; ok {
			return HTML(found)
		}
		return HTML(formatType)
	}

	// deprecated link formats
	return handleDeprecatedLink(b, formatType, o)
}

func handleVariable(b block, formatType string, o *FmtOpt) HTML {
	// fetch the value
	val, errHTML := getVariableValue(b, formatType, o)
	if errHTML != "" {
		return errHTML
	}

	return formatVariableValue(b, formatType, val, o)
}

func getVariableValue(b block, formatType string, o *FmtOpt) (any, HTML) {
	val, err := b.variables().Get(formatType[1:])
	if err != nil {
		if !o.NoWarnings {
			b.warn(o.Pos, err.Error())
		}
		return nil, HTML("(error: " + formatType + ": " + html.EscapeString(err.Error()) + ")")
	}
	if val == nil {
		if !o.NoWarnings {
			b.warn(o.Pos, "variable "+formatType+" is undefined (in "+b.blockType()+"{} scope)")
		}
		return nil, HTML("(null)")
	}
	return val, ""
}

func formatVariableValue(b block, formatType string, val any, o *FmtOpt) HTML {
	strVal, isStr := val.(string)
	htmlVal, isHTML := val.(HTML)

	// %var is for unformatted strings only
	if formatType[0] == '%' {
		return formatUnescapedVariable(b, formatType, strVal, htmlVal, isStr, isHTML, o)
	}

	// @var escapes strings but preserves HTML
	if isStr {
		return HTML(html.EscapeString(strVal))
	}
	if isHTML {
		return htmlVal
	}
	return HTML(html.EscapeString(humanReadableValue(val)))
}

func formatUnescapedVariable(b block, formatType string, strVal string, htmlVal HTML, isStr, isHTML bool, o *FmtOpt) HTML {
	if isHTML {
		// warn that HTML is being double-encoded
		b.warn(o.Pos, "variable "+formatType+" already formatted; use @"+formatType[1:])
		return htmlVal
	} else if !isStr {
		// other non-string value, probably a block
		b.warn(o.Pos, "can't interpolate non-string variable "+formatType)
		return HTML("(error: " + formatType + ": interpolating non-string)")
	}
	// recursively format the string but prevent infinite variable interpolation
	return formatOpts(b, strVal, o.Pos, FmtOpt{noVariables: true})
}

func handleLink(b block, formatType string, o *FmtOpt) HTML {
	// parse the link and generate HTML
	ok, target, linkType, tooltip, display := parseLink(b, formatType[1:len(formatType)-1], o)
	invalid := ""
	if !ok {
		invalid = " invalid"
	}
	if tooltip != "" {
		tooltip = ` title="` + tooltip + `"`
	}
	return HTML(fmt.Sprintf(`<a class="q-link-%s%s" href="%s"%s>%s</a>`,
		linkType, invalid, target, tooltip, display))
}

func handleDeprecatedLink(b block, formatType string, o *FmtOpt) HTML {
	// handle old !link!, ~link~, $link$ formats for backwards compatibility
	if !oldLinkRegex.MatchString(formatType) {
		return HTML("")
	}

	match := oldLinkRegex.FindStringSubmatch(formatType)
	linkChar, inner := match[1], match[2]
	text, target := inner, inner

	// split on last | for text|target format
	if pipe := strings.LastIndexByte(inner, '|'); pipe != -1 {
		text = inner[:pipe]
		target = inner[pipe+1:]
	}

	// convert to new link format based on prefix character
	switch linkChar[0] {
	case '!':
		formatType = text + "|wp:" + target // wikipedia link
	case '~':
		formatType = text + "|~" + target // category link
	case '$':
		formatType = text + "|" + target // external link
	}

	return handleLink(b, "["+formatType+"]", o)
}

func parseLink(b block, link string, o *FmtOpt) (ok bool, target, linkType, tooltip string, display HTML) {
	ok = true
	// let's not waste any time here
	if link == "" {
		return
	}

	// parse display|target format and determine link type
	display, target, displayDefault := parseLinkComponents(b, link, o)
	tooltip = target

	linkType, target, tooltip, displayDefault, handler := determineLinkType(b, target, displayDefault)

	// let page options modify the link
	if handler != nil {
		handler(b.page(), &PageOptLinkOpts{
			Ok:             &ok,
			Target:         &target,
			Tooltip:        &tooltip,
			DisplayDefault: &displayDefault,
			FmtOpt:         o,
		})
	}

	// use default display if none specified
	if display == "" {
		display = HTML(html.EscapeString(displayDefault))
	}

	target = strings.TrimSpace(target)
	tooltip = strings.TrimSpace(tooltip)
	return
}

func parseLinkComponents(b block, link string, o *FmtOpt) (display HTML, target, displayDefault string) {
	// split on | to separate display text from target
	split := strings.SplitN(link, "|", 2)
	if len(split) == 2 {
		display = format(b, strings.TrimSpace(split[0]), o.Pos)
		target = strings.TrimSpace(split[1])
	} else {
		target = strings.TrimSpace(split[0])
	}
	displayDefault = target
	return
}

func determineLinkType(b block, target, displayDefault string) (linkType, newTarget, tooltip, newDisplayDefault string, handler PageOptLinkFunction) {
	// determine what kind of link this is based on target format
	switch {
	case linkRegex.MatchString(target):
		return parseOtherLink(target)

	case strings.HasPrefix(target, "mailto:"):
		return parseMailtoLink(target)

	case mailRegex.MatchString(target):
		return parseEmailLink(target)

	case wikiRegex.MatchString(target):
		return parseWikiLink(b, target)

	case strings.HasPrefix(target, "~"):
		return parseCategoryLink(b, target)

	default:
		// assume internal page link
		newTarget, tooltip = parseInternalLink(b, target)
		return "internal", newTarget, tooltip, displayDefault, b.page().Opt.Link.ParseInternal
	}
}

func parseOtherLink(target string) (linkType, newTarget, tooltip, displayDefault string, handler PageOptLinkFunction) {
	// external URL or $prefixed link
	linkType = "other"
	displayDefault = linkRegex.ReplaceAllString(target, "")
	if target[0] == '$' {
		target = displayDefault // strip the $ prefix
	}
	return linkType, target, "", displayDefault, nil
}

func parseMailtoLink(target string) (linkType, newTarget, tooltip, displayDefault string, handler PageOptLinkFunction) {
	email := strings.TrimPrefix(target, "mailto:")
	return "contact", target, "email " + email, email, nil
}

func parseEmailLink(target string) (linkType, newTarget, tooltip, displayDefault string, handler PageOptLinkFunction) {
	return "contact", "mailto:" + target, "email " + target, target, nil
}

func parseWikiLink(b block, target string) (linkType, newTarget, tooltip, displayDefault string, handler PageOptLinkFunction) {
	// wiki:target format for external wikis
	matches := wikiRegex.FindStringSubmatch(target)
	tooltip = strings.TrimSpace(matches[1])   // wiki name
	newTarget = strings.TrimSpace(matches[2]) // page name
	displayDefault = newTarget
	handler = b.page().Opt.Link.ParseExternal
	if handler == nil {
		handler = defaultExternalLink
	}
	return "external", newTarget, tooltip, displayDefault, handler
}

func parseCategoryLink(b block, target string) (linkType, newTarget, tooltip, displayDefault string, handler PageOptLinkFunction) {
	// ~category format for category links
	tooltip = strings.TrimPrefix(target, "~")
	newTarget = b.page().Opt.Root.Category + "/" + CategoryNameNE(tooltip)
	return "category", newTarget, tooltip, tooltip, b.page().Opt.Link.ParseCategory
}

func parseInternalLink(b block, target string) (string, string) {
	// handle internal page links with optional sections
	pfx := ""
	if target[0] == '/' {
		target = target[1:] // absolute path
	} else {
		// relative to current page prefix
		pfx = b.page().Prefix()
		if pfx != "" {
			pfx += "/"
		}
	}

	sec := ""
	tooltip := target
	// check for #section syntax
	if hashIdx := strings.IndexByte(target, '#'); hashIdx != -1 {
		sec = PageNameNE(strings.TrimSpace(target[hashIdx+1:]))
		target = strings.TrimSpace(target[:hashIdx])
		tooltip = target + " § " + sec
		sec = "#" + sec
	}

	// handle section-only links
	if target == "" && sec != "" {
		return sec, tooltip
	}

	return b.page().Opt.Root.Page + "/" + pfx + PageNameNE(target) + sec, tooltip
}

func defaultExternalLink(p *Page, o *PageOptLinkOpts) {
	// handle external wiki links like wp:Article_Name
	// note: the wiki shortcode is in tooltip for now
	// the target is in displayDefault
	ext, exists := p.Opt.External[*o.Tooltip]
	if !exists {
		p.warn(o.Pos, "external wiki '"+*o.Tooltip+"' does not exist")
		*o.Ok = false
		return
	}

	// default tooltip for no section
	*o.Tooltip = ext.Name + ": " + *o.Target // e.g. wikipedia: some page

	// split by # to get section
	section := ""
	split := strings.SplitN(*o.Target, "#", 2)
	if len(split) == 2 {
		*o.Target = strings.TrimSpace(split[0])
		section = strings.TrimSpace(split[1])
		*o.Tooltip = *o.Target + " § " + section
	}

	// normalize based on type
	switch ext.Type {

	// convert all non-alphanumerics to underscore
	case PageOptExternalTypeQuiki:
		*o.Target = PageNameLink(*o.Target)
		section = PageNameLink(section)

	// convert space to underscore, URI escape the rest
	case PageOptExternalTypeMediaWiki:
		*o.Target = html.EscapeString(strings.Replace(*o.Target, " ", "_", -1))
		section = html.EscapeString(strings.Replace(section, " ", "_", -1))

	// no special normalization, just URI escapes
	default: // (PageOptExternalTypeNone)
		*o.Target = html.EscapeString(*o.Target)
		section = html.EscapeString(*o.Target)
	}

	// add the wiki page root
	*o.Target = ext.Root + "/" + *o.Target

	// add the section back
	if section != "" {
		*o.Target += "#" + section
	}
}
