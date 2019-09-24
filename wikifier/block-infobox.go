package wikifier

type infobox struct {
	*Map
}

func newInfobox(name string, b *parserBlock) block {
	b.typ = "infobox"
	return &infobox{newMapBlock("", b).(*Map)}
}

func (ib *infobox) parse(page *Page) {
	ib.Map.parse(page)
}

func (ib *infobox) html(page *Page, el element) {
	ib.Map.html(page, nil)
	el.setTag("table")

	// display the title if there is one
	if ib.name != "" {
		th := el.createChild("tr", "infobox-title").createChild("th", "")
		th.setAttr("colspan", "2")
		th.addHtml(page.parseFormattedTextOpts(ib.name, &formatterOptions{pos: ib.openPos}))
	}

	// add the rows
	infoboxTableAddRows(ib, el, page, ib.mapList)
}

// # append each pair.
// # note that $table might actually be a Wikifier::Elements container
// sub table_add_rows {
func infoboxTableAddRows(infoboxOrSec block, table element, page *Page, pairs []mapListEntry) {
	//     my @pairs = $block->map_array;
	//     my $has_title = 0;
	hasTitle := false

	//     for (0..$#pairs)
	for i, entry := range pairs {

		// if the value is from infosec{}, add each row
		if els, ok := entry.value.(element); ok && els.hasMeta("infosec") {

			// infosec do not need a key
			if entry.keyTitle != "" {
				infoboxOrSec.warn(infoboxOrSec.openPosition(), "Key associated with infosec{} ignored")
			}

			table.addChild(els)
		}

		// determine next entry
		var next mapListEntry
		hasNext := i != len(pairs)-1
		if hasNext {
			next = pairs[i+1]
		}

		//         # options based on position in the infosec
		//         my @classes;
		//         push @classes, 'infosec-title' and $has_title++ if $is_title;
		//         push @classes, 'infosec-first' if $_ == $has_title;
		//         my $b4_infosec = $next && blessed $next->{value} &&
		//             $next->{value}{is_infosec};
		//         push @classes, 'infosec-last'
		//             if !$is_title && ($b4_infosec || $_ == $#pairs);

		//         my %row_opts = (
		//             is_block => $is_block,
		//             is_title => $is_title,
		//             td_opts  => { classes => \@classes }
		//         );

		//         # not an infosec{}; this is a top-level pair
		//         table_add_row($table, $page, $key_title, $value, $pos, \%row_opts);
		//     }
		// }
	}
}

// # add a row.
// # note that $table might actually be a Wikifier::Elements container
// sub table_add_row {
//     my ($table, $page, $key_title, $value, $pos, $opts_) = @_;
//     my %opts    = hash_maybe $opts_;
//     my %td_opts = hash_maybe $opts{td_opts};

//     # create the row.
//     my $tr = $table->create_child(
//         type  => 'tr',
//         class => 'infobox-pair'
//     );

//     # append table row with key.
//     if (length $key_title) {
//         $key_title = $page->parse_formatted_text($key_title, pos => $pos);
//         $tr->create_child(
//             type       => 'th',
//             class      => 'infobox-key',
//             content    => $key_title,
//             %td_opts
//         );
//         $tr->create_child(
//             type       => 'td',
//             class      => 'infobox-value',
//             content    => $value,
//             %td_opts
//         );
//     }

//     # append table row without key.
//     else {
//         my $td = $tr->create_child(
//             type       => 'td',
//             class      => 'infobox-anon',
//             attributes => { colspan => 2 },
//             content    => $value,
//             %td_opts
//         );
//         $td->add_class('infobox-text') if !$opts{is_block};
//     }
// }
