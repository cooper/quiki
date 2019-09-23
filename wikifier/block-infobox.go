package wikifier

type infobox struct {
	*Map
	*parserBlock
}

func newInfobox(name string, b *parserBlock) block {
	// FIXME: this code to create the underlying map is pretty messy
	return &infobox{newMapBlock("", &parserBlock{
		openPos:      b.openPos,
		parentB:      b.parentB,
		parentC:      b.parentC,
		typ:          "map",
		genericCatch: &genericCatch{},
	}).(*Map), b}
}

func (ib *infobox) parse(page *Page) {
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
	infoboxTableAddRows(el, page, ib.mapList)
}

// # append each pair.
// # note that $table might actually be a Wikifier::Elements container
// sub table_add_rows {
func infoboxTableAddRows(table element, page *Page, pairs []mapListEntry) {
	//     my @pairs = $block->map_array;
	//     my $has_title = 0;
	hasTitle := false

	//     for (0..$#pairs)
	for i, entry := range pairs {

		//         my ($key_title, $value, $key, $pos, $is_block, $is_title) =
		//             @{ $pairs[$_] }{ qw(key_title value key pos is_block is_title) };
		//         my $next = $pairs[$_ + 1];

		//         # if the value is from infosec{}, add each row
		//         if (blessed $value && $value->{is_infosec}) {
		//             #warning("Key associated with infosec{} ignored")
		//             #    if length $key_title;
		//             $table->add($value);
		//             next;
		//		   }d
		if els, ok := entry.value.(element); ok && els.hasMeta("infosec") {

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
