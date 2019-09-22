package wikifier

import (
	"fmt"
	htmlfmt "html"
	"regexp"
	"strings"
)

var variableRegex = regexp.MustCompile(`^([@%])([\w\.]+)$`)
var linkRegex = regexp.MustCompile(`^((\w+)://|\$)`)
var mailRegex = regexp.MustCompile(`^[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,63}$`)
var colorRegex = regexp.MustCompile(`(?i)^#[\da-f]+$`)
var wikiRegex = regexp.MustCompile(`^(\w+):`)

var linkNormalizers = map[string]func(string) string{
	"wikifier": func(s string) string {
		return pageNameLink(s)
	},
	"mediawiki": func(s string) string {
		s = strings.Replace(s, " ", "_", -1)
		return htmlfmt.EscapeString(s)
	},
	"none": func(s string) string {
		return htmlfmt.EscapeString(s)
	},
}

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
	"i":   `<span style="font-style: italic;">`,            // italic
	"b":   `<span style="font-weight: bold;">`,             // bold
	"s":   `<span style="text-decoration: line-through;">`, // strike
	"c":   `<code>`,                                        // inline code
	"/c":  `</code>`,
	"/s":  `</span>`,
	"/b":  `</span>`,
	"/i":  `</span>`,
	"q":   `<span style="font-style: italic;">"`, // inline quote
	"/q":  `"</span>`,
	"^":   `<sup>`, // superscript
	"/^":  `</sup>`,
	"v":   `<sub>`, // subscript
	"/v":  `</sub>`,
	"/":   `</span>`,
	"nl":  `<br />`,  // line break
	"br":  `<br />`,  // (deprecated)
	"--":  `&ndash;`, // en dash
	"---": `&mdash;`, // em dash
}

type formatterOptions struct {
	noEntities  bool     // disables html entity conversion
	noVariables bool     // used internally to prevent recursive interpolation
	noWarnings  bool     // silence warnings for undefined variables
	pos         position // position used for warnings
	startPos    position // set internally to position of '['
}

func (page *Page) parseFormattedText(text string) Html {
	return page.parseFormattedTextOpts(text, &formatterOptions{})
}

func (page *Page) parseFormattedTextOpts(text string, opts *formatterOptions) Html {

	// let's not waste any time here
	if text == "" {
		return ""
	}

	// find and copy the position
	if opts.pos.none() {
		// TODO: use the current page position
	}

	// my @items;
	var items []interface{} // string and html
	str := ""
	formatType := "" // format name such as 'i' or '/b'
	formatDepth := 0 // how far [[in]] we are
	escaped := false // character escaped

	for _, char := range text {

		// update position
		if char == '\n' {
			opts.pos.line++
			opts.pos.column = 0
		} else {
			opts.pos.column++
		}

		if char == '[' && !escaped {
			// marks the beginning of a formatting element
			formatDepth++
			if formatDepth == 1 {
				opts.startPos = opts.pos
				formatType = ""

				// store the string we have so far
				if str != "" {
					if opts.noEntities {
						items = append(items, Html(str))
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
				items = append(items, page.parseFormatType(formatType, opts))
				opts.startPos = position{}
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
		if opts.noEntities {
			items = append(items, Html(str))
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
			final += htmlfmt.EscapeString(v)
		case Html:
			final += string(v)
		}
	}

	return Html(final)
}

func (page *Page) parseFormatType(formatType string, opts *formatterOptions) Html {

	// static format
	if format, exists := staticFormats[formatType]; exists {
		return Html(format)
	}

	// variable
	if !opts.noVariables {
		if variableRegex.MatchString(formatType) {
			// TODO: fetch the variable, warn if it is undef, format text if %var
			// then return the value
			return Html("(null)")
		}
	}

	// # html entity.
	if formatType[0] == '&' {
		return Html("&" + formatType[1:] + ";")
	}

	// # deprecated: a link in the form of [~link~], [!link!], or [$link$]
	// # convert to newer link format
	// if ($type =~ /^([\!\$\~]+?)(.+)([\!\$\~]+?)$/) {
	//     my ($link_char, $inner) = ($1, $2);
	//     my ($target, $text) = ($inner, $inner);

	//     # format is <text>|<target>
	//     if ($inner =~ m/^(.+)\|(.+?)$/) {
	//         $text   = $1;
	//         $target = $2;
	//     }

	//     # category wiki link [~ category ~]
	//     if ($link_char eq '~') {
	//         $type = "[ $text | ~ $target ]";
	//     }

	//     # external wiki link [! article !]
	//     # technically this used to observe @external.name and @external.root,
	//     # but in practice this was always set to wikipedia, so use 'wp'
	//     elsif ($link_char eq '!') {
	//         $type = "[ $text | wp: $target ]";
	//     }

	//     # other non-wiki link [$ url $]
	//     elsif ($link_char eq '$') {
	//         $type = "[ $text | $target ]";
	//     }
	// }

	// [[link]]
	if formatType[0] == '[' && formatType[len(formatType)-1] == ']' {
		ok, displaySame, target, display, tooltip, linkType := parseLink(formatType[1 : len(formatType)-2])
		invalid := ""
		if !ok {
			invalid = " invalid"
		}
		if !displaySame {
			display = string(page.parseFormattedText(display))
		}
		return Html(fmt.Sprintf(`<a class="q-link-%s%s" href="%s"%s>%s</a>`,
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
		return Html(`<span style="color: "` + color + `";">`)
	}

	// color hex code
	if colorRegex.MatchString(formatType) {
		return Html(`<span style="color: "` + formatType + `";">`)
	}

	// TODO: real references.
	// if ($type =~ m/^\d+$/) {
	//     return qq{<sup style="font-size: 75%"><a href="#wiki-ref-$type" class="wiki-ref-anchor">[$type]</a></sup>};
	// }

	return Html("")
}

func parseLink(link string) (ok, displaySame bool, target, display, tooltip, linkType string) {
	var normalizer func(ok *bool, target, tooltip, display *string)

	// split into display and target
	split := strings.SplitN(link, "|", 2)
	display = strings.TrimSpace(split[0])
	if len(split) == 2 {
		target = strings.TrimSpace(split[1])
	}

	// no pipe
	if target == "" {
		target = display
		displaySame = true
	}

	if linkRegex.MatchString(target) {
		// http://google.com or $/something (see wikifier issue #68)

		linkType = "other"
		normalizer = normalizeOtherLink

		// erase the scheme or $
		if displaySame {
			display = linkRegex.ReplaceAllString(display, "")
		}
	} else if strings.HasPrefix(target, "mailto:") {
		// mailto:someone@example.com

		linkType = "contact"
		normalizer = normalizeEmailLink
		if displaySame {
			display = strings.TrimPrefix(target, "mailto:")
		}

	} else if mailRegex.MatchString(target) {
		// someone @example.com

		linkType = "contact"
		normalizer = normalizeEmailLink

	} else if s := wikiRegex.FindStringSubmatch(target); len(s) != 0 {
		// wp: some page

		// FIXME: I think s[0] is wp
		linkType = "external"
		normalizer = normalizeExternalLink

		if displaySame {
			display = wikiRegex.ReplaceAllString(target, "")
		}

	} else if strings.HasPrefix(target, "~") {
		// ~ some category

		target = strings.TrimPrefix(target, "~")
		target = strings.TrimSpace(target)
		normalizer = normalizeCategoryLink

		if displaySame {
			display = target
		}
	} else {
		// normal page link

		linkType = "internal"
		normalizer = normalizePageLink
	}

	// normalize
	target = strings.TrimSpace(target)
	tooltip = strings.TrimSpace(tooltip)
	display = strings.TrimSpace(display)

	normalizer(&ok, &target, &tooltip, &display)
	return
}

func normalizeEmailLink(ok *bool, target, tooltip, display *string) {
	email := *target
	if strings.HasPrefix(email, "mailto:") {
		email = strings.TrimPrefix(email, "mailto:")
	} else {
		*target = "mailto:" + email
	}
	*tooltip = "Email " + email
}

func normalizeOtherLink(ok *bool, target, tooltip, display *string) {
	*target = strings.TrimPrefix(*target, "$")
	*ok = true
}

func normalizeCategoryLink(ok *bool, target, tooltip, display *string) {
	//TODO
}

func normalizePageLink(ok *bool, target, tooltip, display *string) {
	//TODO

}

func normalizeExternalLink(ok *bool, target, tooltip, display *string) {
	//TODO

}
