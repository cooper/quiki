package wikifier

import (
	htmlfmt "html"
	"regexp"
	"strings"
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

func parseFormattedText(text string) html {
	return parseFormattedTextOpts(text, &formatterOptions{})
}

func parseFormattedTextOpts(text string, opts *formatterOptions) html {

	// let's not waste any time here
	if text == "" {
		return ""
	}

	// find and copy the position
	if opts.pos.line == 0 && opts.pos.column == 0 {
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
						items = append(items, html(str))
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
				items = append(items, parseFormatType(formatType, opts))
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
			items = append(items, html(str))
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
		case html:
			final += string(v)
		}
	}

	return html(final)
}

func parseFormatType(formatType string, opts *formatterOptions) html {

	// static format
	if format, exists := staticFormats[formatType]; exists {
		return html(format)
	}

	// variable
	if !opts.noVariables {
		if match, _ := regexp.MatchString(`^([@%])([\w\.]+)$`, formatType); match {
			// TODO: fetch the variable, warn if it is undef, format text if %var
			// then return the value
			return html("(null)")
		}
	}

	// # html entity.
	if formatType[0] == '&' {
		return html("&" + formatType[1:] + ";")
	}

	// TODO: links, references

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

	// # [[link]]
	// if ($type =~ /^\[(.+)\]$/) {
	//     my ($ok, $target, $display, $tooltip, $link_type, $display_same) =
	//         $wikifier->parse_link($page, $1, %opts);

	//     # text formatting is permitted before the pipe.
	//     # do nothing when the link did not have a pipe ($display_same)
	//     $display = $wikifier->parse_formatted_text($page, $display, %opts)
	//         unless $display_same;

	//     return sprintf '<a class="wiki-link-%s%s" href="%s"%s>%s</a>',
	//         $link_type,
	//         $ok ? '' : ' invalid',
	//         $target,
	//         length $tooltip ? qq{ title="$tooltip"} : '',
	//         $display;
	// }

	// # fake references.
	// if ($type eq 'ref') {
	//     $page->{reference_number} ||= 1;
	//     my $ref = $page->{reference_number}++;
	//     return qq{<sup style="font-size: 75%"><a href="#wiki-ref-$ref" class="wiki-ref-anchor">[$ref]</a></sup>};
	// }

	// # color name.
	// if (my $color = $colors{ lc $type }) {
	//     return qq{<span style="color: $color;">};
	// }

	// color name
	if color, exists := colors[strings.ToLower(formatType)]; exists {
		return html(`<span style="color: "` + color + `";">`)
	}

	// color hex code
	if match, _ := regexp.MatchString(`(?i)^#[\da-f]+$`, formatType); match {
		return html(`<span style="color: "` + formatType + `";">`)
	}

	// # real references.
	// if ($type =~ m/^\d+$/) {
	//     return qq{<sup style="font-size: 75%"><a href="#wiki-ref-$type" class="wiki-ref-anchor">[$type]</a></sup>};
	// }

	return html("")

}
