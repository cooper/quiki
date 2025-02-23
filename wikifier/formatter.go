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

// Fmt generates HTML from a quiki-encoded formatted string.
func (b *parserBlock) Fmt(text string, pos Position) HTML {
	return b._parseFormattedText(text, &FmtOpt{Pos: pos})
}

// FmtOpts is like Fmt except you can specify additional options with the FmtOpt argument.
func (b *parserBlock) FmtOpts(text string, pos Position, o FmtOpt) HTML {
	o.Pos = pos
	return b._parseFormattedText(text, &o)
}

func (b *parserBlock) _parseFormattedText(text string, o *FmtOpt) HTML {

	// let's not waste any time here
	if text == "" {
		return ""
	}

	// find and copy the position
	if o.Pos.none() && b._page.parser != nil {
		o.Pos = b._page.parser.pos
	}

	// my @items;
	var items []any // string and html
	str := ""
	formatType := "" // format name such as 'i' or '/b'
	formatDepth := 0 // how far [[in]] we are
	escaped := false // character escaped

	for _, char := range text {

		// update position
		if char == '\n' {
			o.Pos.Line++
			o.Pos.Column = 0
		} else {
			o.Pos.Column++
		}

		if char == '[' && !escaped {
			// marks the beginning of a formatting element
			formatDepth++
			if formatDepth == 1 {
				formatType = ""

				// store the string we have so far
				if str != "" {
					if o.NoEntities {
						items = append(items, HTML(str))
					} else {
						items = append(items, str)
					}
					str = ""
				}

				continue
			}
		} else if char == ']' && !escaped && formatDepth != 0 {
			// marks the end of a formatting element
			formatDepth--
			if formatDepth == 0 {
				items = append(items, b.parseFormatType(formatType, o))
				continue
			}
		}

		// an unescaped backslash should not appear in the result
		escaped = char == '\\' && !escaped
		if escaped && formatDepth == 0 {
			continue
		}

		// if we're in the format type, append to it
		if formatDepth != 0 {
			formatType += string(char)
		} else {
			// otherwise, add to the string
			str += string(char)
		}
	}

	// add the final string
	if str != "" {
		if o.NoEntities {
			items = append(items, HTML(str))
		} else {
			items = append(items, str)
		}
	}

	// TODO: this could be a block
	// # might be a blessed object
	// return $items[0][1] if $#items == 0 && blessed $items[0][1];

	// join the parts together, converting entities as needed
	final := ""
	for _, piece := range items {
		switch v := piece.(type) {
		case string:
			final += html.EscapeString(v)
		case HTML:
			final += string(v)
		}
	}

	return HTML(final)
}

func (b *parserBlock) parseFormatType(formatType string, o *FmtOpt) HTML {

	// static format
	if format, exists := staticFormats[strings.ToLower(formatType)]; exists {
		return HTML(format)
	}

	// variable
	if !o.noVariables {
		if variableRegex.MatchString(formatType) {

			// fetch the value
			val, err := b.variables().Get(formatType[1:])
			if err != nil {
				if !o.NoWarnings {
					b.warn(o.Pos, err.Error())
				}
				return HTML("(error: " + formatType + ": " + html.EscapeString(err.Error()) + ")")
			}
			if val == nil {
				if !o.NoWarnings {
					b.warn(o.Pos, "Variable "+formatType+" is undefined")
				}
				return HTML("(null)")
			}

			strVal, isStr := val.(string)
			htmlVal, isHTML := val.(HTML)

			// %var is for unformatted strings only
			if formatType[0] == '%' {
				if isHTML {
					// warn that HTML is being double-encoded
					b.warn(o.Pos, "Variable "+formatType+" already formatted; use @"+formatType[1:])
					return htmlVal
				} else if !isStr {
					// other non-string value, probably a block
					b.warn(o.Pos, "Can't interpolate non-string variable "+formatType)
					return HTML("(error: " + formatType + ": interpolating non-string)")
				}
				return b.FmtOpts(strVal, o.Pos, FmtOpt{noVariables: true})
			}

			// @var with a string (was set with %var but should not be interpolated)
			if isStr {
				return HTML(html.EscapeString(strVal))
			}

			// @var with HTML (normal set with @ and retrieved with @)
			if isHTML {
				return htmlVal
			}

			// TODO: if it's a block, maybe we should complain to use {@var} syntax

			// I don't really know what to do
			return HTML(html.EscapeString(humanReadableValue(val)))
		}
	}

	// # html entity.
	if formatType[0] == '&' {
		return HTML("&" + formatType[1:] + ";")
	}

	// # deprecated: a link in the form of [~link~], [!link!], or [$link$]
	// # convert to newer link format
	if formatType[0] != '[' {
		if match := oldLinkRegex.FindStringSubmatch(formatType); match != nil {
			linkChar, inner := match[1], match[2]
			text, target := inner, inner

			// format is <text>|<target>
			if pipe := strings.LastIndexByte(inner, '|'); pipe != -1 {
				text = inner[:pipe]
				target = inner[pipe+1:]
			}

			switch linkChar[0] {

			// external wiki link
			// technically this used to observe @external.name and @external.root,
			// but in practice it was always set to wikipedia
			case '!':
				formatType = text + "|wp:" + target

			// category link
			case '~':
				formatType = text + "|~" + target

			// other non-wiki link
			case '$':
				formatType = text + "|" + target

			}

			formatType = "[" + formatType + "]"
		}
	}

	// [[link]]
	if formatType[0] == '[' && formatType[len(formatType)-1] == ']' {
		ok, target, linkType, tooltip, display := b.parseLink(formatType[1:len(formatType)-1], o)
		invalid := ""
		if !ok {
			invalid = " invalid"
		}
		if tooltip != "" {
			tooltip = ` title="` + tooltip + `"`
		}
		return HTML(fmt.Sprintf(`<a class="q-link-%s%s" href="%s"%s>%s</a>`,
			linkType,
			invalid,
			target,
			tooltip,
			display,
		))
	}

	// TODO: fake references.
	// if ($type eq 'ref') {
	//     $page->{reference_number} ||= 1;
	//     my $ref = $page->{reference_number}++;
	//     return qq{<sup style="font-size: 75%"><a href="#wiki-ref-$ref" class="wiki-ref-anchor">[$ref]</a></sup>};
	// }

	// color name
	if color, exists := colors[strings.ToLower(formatType)]; exists {
		return HTML(`<span style="color: "` + color + `";">`)
	}

	// color hex code
	if colorRegex.MatchString(formatType) {
		return HTML(`<span style="color: "` + formatType + `";">`)
	}

	// TODO: real references.
	// if ($type =~ m/^\d+$/) {
	//     return qq{<sup style="font-size: 75%"><a href="#wiki-ref-$type" class="wiki-ref-anchor">[$type]</a></sup>};
	// }

	// inline html
	// [html:x<sup>2</sup>]
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

	return HTML("")
}

func (b *parserBlock) parseLink(link string, o *FmtOpt) (ok bool, target, linkType, tooltip string, display HTML) {
	ok = true

	// nothing in, nothing out
	if link == "" {
		return
	}

	// split into display and target
	split := strings.SplitN(link, "|", 2)
	displayDefault := ""
	if len(split) == 2 {
		display = b.Fmt(strings.TrimSpace(split[0]), o.Pos)
		target = strings.TrimSpace(split[1])
	} else {
		target = strings.TrimSpace(split[0])
	}
	tooltip = target
	displayDefault = target

	var handler PageOptLinkFunction
	if matches := linkRegex.FindStringSubmatch(target); len(matches) != 0 {
		// http://google.com or $/something (see wikifier issue #68)

		linkType = "other"

		// tooltip can't really add any value
		tooltip = ""

		// erase the scheme or $ from display
		displayDefault = linkRegex.ReplaceAllString(target, "")

		// erase the $ from target
		if target[0] == '$' {
			target = displayDefault
		}

	} else if strings.HasPrefix(target, "mailto:") {
		// mailto:someone@example.com

		linkType = "contact"
		email := strings.TrimPrefix(target, "mailto:")
		tooltip = "Email " + email

		// erase mailto:
		displayDefault = email

	} else if mailRegex.MatchString(target) {
		// someone@example.com

		linkType = "contact"
		tooltip = "Email " + target
		target = "mailto:" + target

	} else if s := wikiRegex.FindStringSubmatch(target); len(s) != 0 {
		// wp: some page
		linkType = "external"
		tooltip = strings.TrimSpace(s[1]) // for now
		target = strings.TrimSpace(s[2])  // for now
		displayDefault = target
		handler = b._page.Opt.Link.ParseExternal
		if handler == nil {
			handler = defaultExternalLink
		}

	} else if strings.HasPrefix(target, "~") {
		// ~ some category
		linkType = "category"
		tooltip = strings.TrimPrefix(target, "~")
		target = b._page.Opt.Root.Category + "/" + CategoryNameNE(tooltip)
		displayDefault = tooltip
		handler = b._page.Opt.Link.ParseCategory

	} else {
		// normal page link
		linkType = "internal"

		// prefix
		pfx := ""
		if target[0] == '/' && len(target) > 1 {
			// if target starts with /, ignore prefix
			target = target[1:]
		} else {
			// determine page prefix
			// TODO: resolve . and .. safely
			pfx = b._page.Prefix()
			if pfx != "" {
				pfx += "/"
			}
		}

		// section
		sec := ""
		if hashIdx := strings.IndexByte(target, '#'); hashIdx != -1 && len(target) >= hashIdx {
			sec = PageNameNE(strings.TrimSpace(target[hashIdx+1:]))
			target = strings.TrimSpace(target[:hashIdx])
			tooltip = target + " § " + sec
			sec = "#" + sec
		}

		// determine actual target
		if target == "" && sec != "" {
			// section on same page
			target = sec
		} else {
			// other page link
			target = b._page.Opt.Root.Page + "/" + pfx + PageNameNE(target) + sec
		}

		handler = b._page.Opt.Link.ParseInternal
	}

	// call link handler
	if handler != nil {
		handler(b._page, &PageOptLinkOpts{
			Ok:             &ok,
			Target:         &target,
			Tooltip:        &tooltip,
			DisplayDefault: &displayDefault,
			FmtOpt:         o,
		})
	}

	// pipe was not present
	if display == "" {
		display = HTML(html.EscapeString(displayDefault))
	}

	// normalize
	target = strings.TrimSpace(target)
	tooltip = strings.TrimSpace(tooltip)

	return
}

func defaultExternalLink(p *Page, o *PageOptLinkOpts) {
	// note: the wiki shortcode is in tooltip for now
	// the target is in displayDefault
	ext, exists := p.Opt.External[*o.Tooltip]
	if !exists {

		p.warn(o.Pos, "External wiki '"+*o.Tooltip+"' does not exist")
		*o.Ok = false
		return
	}

	// default tooltip for no section
	*o.Tooltip = ext.Name + ": " + *o.Target // e.g. Wikipedia: Some Page

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
