package wikifier

import (
	"regexp"
	"strconv"
	"strings"
)

var whitespaceRegex = regexp.MustCompile(`\s`)

type styleBlock struct {
	style styleEntry
	*Map
}

type styleEntry struct {
	parent        element    // the element the style{} block belongs to
	applyToParent bool       // apply the rules to the parent itself
	applyTo       [][]string // matcher sets applied to descendants of the parent
	rules         map[string]string
}

func newStyleBlock(name string, b *parserBlock) block {
	return &styleBlock{styleEntry{}, newMapBlock("", b).(*Map)}
}

func (sb *styleBlock) parse(page *Page) {

	// my ($block, $page) = (shift, @_);

	// # parse the hash.
	// $block->{no_format_values}++;
	// $block->parse_base(@_);
	sb.noFormatValues = true
	sb.Map.parse(page)

	// # find rules
	// my %rules;
	// foreach my $item ($block->map_array) {
	//     $rules{ $item->{key_title} } = $item->{value};
	// }

	rules := make(map[string]string, len(sb.mapList))
	for _, entry := range sb.mapList {
		if str, ok := entry.value.(string); ok {
			rules[entry.keyTitle] = str
		} else {
			sb.warn(entry.pos, "Only string values are allowed")
		}
	}

	// # create style
	// my %style = (
	//     apply_to => [],
	//     rules    => \%rules
	// );

	style := styleEntry{rules: rules}

	// # if the block has a name, it applies to child class(es).
	// if (length $block->name) {
	if blockName := sb.blockName(); blockName != "" {
		//     my @matchers = split ',', $block->name;
		//     foreach my $matcher (@matchers) {
		for _, matcher := range strings.Split(blockName, ",") {
			//         $matcher =~ s/^\s*//g;
			//         $matcher =~ s/\s*$//g;
			matcher = strings.TrimSpace(matcher)

			//         # this element.
			//         if ($matcher eq 'this') {
			//             $style{apply_to_parent}++;
			//             next;
			//         }
			if matcher == "this" {
				style.applyToParent = true
				continue
			}

			//         # split up matchers by space.
			//         # replace $blah with model-blah.
			//         my @matchers = split /\s/, $matcher;
			//         @matchers = map { (my $m = $_) =~ s/^\$/model-/; $m } @matchers;

			matchers := whitespaceRegex.Split(matcher, -1)
			for i, m := range matchers {
				if m[0] != '$' {
					continue
				}
				matchers[i] = "model-" + m[1:]
			}

			//         # element type or class, etc.
			//         # ex: p
			//         # ex: p.something
			//         # ex: .something.somethingelse
			//         push @{ $style{apply_to} }, \@matchers;
			style.applyTo = append(style.applyTo, matchers)
		}
	} else {
		style.applyToParent = true
	}

	sb.style = style
}

func (sb *styleBlock) html(page *Page, el element) {
	el.hide()
	// my ($block, $page) = (shift, @_);
	// my %style     = %{ $block->{style} };
	style := sb.style

	// my $parent_el = $block->parent->element;
	parentEl := sb.parentBlock().el()

	// make sure the parent el has an id like pagename-n so the CSS can refer to it.
	// keep an existing id (e.g. the main block's, or one set by a previous style{}
	// on the same parent) so that every style targeting this element agrees on it.
	if parentEl.id() == "" {
		page._styleId++
		parentEl.setId(page.elementIDPrefix() + "-" + strconv.Itoa(page._styleId))
	}

	// the selectors are built from the parent element when the CSS is
	// generated, so that they use whatever id it ends up with
	style.parent = parentEl

	// push @{ $page->{styles} ||= [] }, \%style;
	page.styles = append(page.styles, style)
}
