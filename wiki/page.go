package wiki

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
