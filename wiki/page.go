package wiki

import "errors"

// DisplayPage represents a page result to display.
type DisplayPage struct {

	// basename of the page, with the extension
	File string

	// basename of the page, without the extension
	Name string

	// absolute file path of the page
	Path string

	// the page content (HTML)
	Content string

	// UNIX timestamp of when the page was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the page file itself.
	ModUnix int64

	// like ModUnix except in HTTP date format, suitable for Last-Modified
	Modified string

	// CSS generated for the page from style{} blocks
	CSS string

	// true if this content was read from a cache file. opposite of Generated
	FromCache bool

	// true if the content being served was just generated on the fly.
	// opposite of FromCache
	Generated bool

	// true if this request resulted in the writing of a new cache file.
	// this can only be true if Generated is true
	CacheGenerated bool

	// true if this request resulted in the writing of a text file.
	// this can only be true if Generated is true
	TextGenerated bool

	// true if the page has not yet been published for public viewing.
	// this only occurs when it is specified that serving drafts is OK,
	// since normally a draft page instead results in a DisplayError.
	Draft bool

	// warnings produced by the parser
	Warnings []string

	// UNIX timestamp of when the page was created, as extracted from
	// the special @page.created variable
	CreatedUnix int64

	// name of the page author, as extracted from the special @page.author
	// variable
	Author string

	// list of categories the page belongs to, without the '.cat' extension
	Categories []string

	// page title as extracted from the special @page.title variable, including
	// any possible HTML-encoded formatting
	FmtTitle string

	// like FmtTitle except that all text formatting has been stripped.
	// suitable for use in the <title> tag
	Title string
}

// DisplayPage returns the display result for a page.
func (w *Wiki) DisplayPage(name string) (DisplayPage, error) {
	var result DisplayPage

	// create the page
	page := w.NewPage(name)

	// # get page info
	// my $path = $page->path;
	// my $cache_path = $page->cache_path;

	// file does not exist
	if !page.Exists() {
		return result, errors.New("page does not exist")
	}

	// # filename and path info
	// $result->{file} = $page->name;      # with extension
	// $result->{name} = $page->name_ne;   # without extension
	// $result->{path} = $path;            # absolute path

	// # page content
	// $result->{type} = 'page';
	// $result->{mime} = 'text/html';

	// # FIRST redirect check - this is for symbolic link redirects.
	// return _display_page_redirect($page->redirect, $result)
	//     if $page->is_symlink;

	// # caching is enabled, so let's check for a cached copy.
	// if ($wiki->opt('page.enable.cache') && -f $cache_path) {
	//     $result = $wiki->_get_page_cache($page, $result, \%opts);
	//     return $result if $result->{cached};
	// }

	// # Safe point - we will be generating the page right now.

	// # parse the page.
	// # if an error occurs, parse it again in variable-only mode.
	// # then hopefully we can at least get the metadata and categories.
	// my $err = $page->parse;
	// if ($err) {
	//     $page->{vars_only}++;
	//     $page->parse;
	//     $wiki->cat_check_page($page);
	//     return $wiki->_display_error_cache($page, $err, parse_error => 1);
	// }

	// # if this is a draft, so pretend it doesn't exist
	// if ($page->draft && !$opts{draft_ok}) {
	//     L 'Draft';
	//     return $wiki->_display_error_cache($page,
	//         "Page has not yet been published.",
	//         draft => 1
	//     );
	// }

	// # THIRD redirect check - this is for pages we just generated with
	// # @page.redirect in them.
	// my $redir = _display_page_redirect($page->redirect, $result);
	// return $wiki->_write_page_cache_maybe($page, $redir) if $redir;

	// # generate the HTML and headers.
	// $result->{generated}  = \1;
	// $result->{page}       = $page;
	// $result->{draft}      = $page->draft;
	// $result->{warnings}   = $page->{warnings};
	// $result->{mod_unix}   = time;
	// $result->{modified}   = time2str($result->{mod_unix});
	// $result->{content}    = $page->html;
	// $result->{css}        = $page->css;

	// # update categories. this must come after ->html
	// $wiki->cat_check_page($page);
	// $result->{categories} = [ _cats_to_list($page->{categories}) ];

	// # write cache file if enabled
	// $result = $wiki->_write_page_cache_maybe($page, $result);
	// return $result if $result->{error};

	// # search is enabled, so generate a text file
	// $result = $wiki->_write_page_text($page, $result)
	//     if $wiki->opt('search.enable');

	// return $result;

	return result, nil
}
