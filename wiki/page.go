package wiki

import (
	httpdate "github.com/Songmu/go-httpdate"
	"github.com/cooper/quiki/wikifier"
)

// DisplayPage represents a page result to display.
type DisplayPage struct {

	// basename of the page, with the extension
	File string `json:"file,omitempty"`

	// basename of the page, without the extension
	Name string `json:"name,omitempty"`

	// absolute file path of the page
	Path string `json:"path,omitempty"`

	// the page content (HTML)
	Content wikifier.HTML `json:"content,omitempty"`

	// UNIX timestamp of when the page was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the page file itself.
	ModUnix int64 `json:"mod_unix,omitempty"`

	// like ModUnix except in HTTP date format, suitable for Last-Modified
	Modified string `json:"modified,omitempty"`

	// CSS generated for the page from style{} blocks
	CSS string `json:"css,omitempty"`

	// true if this content was read from a cache file. opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated on the fly.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if this request resulted in the writing of a new cache file.
	// this can only be true if Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`

	// true if this request resulted in the writing of a text file.
	// this can only be true if Generated is true
	TextGenerated bool `json:"text_gen,omitempty"`

	// true if the page has not yet been published for public viewing.
	// this only occurs when it is specified that serving drafts is OK,
	// since normally a draft page instead results in a DisplayError.
	Draft bool `json:"draft,omitempty"`

	// warnings produced by the parser
	Warnings []string `json:"warnings,omitempty"`

	// UNIX timestamp of when the page was created, as extracted from
	// the special @page.created variable
	CreatedUnix int64 `json:"created,omitempty"`

	// name of the page author, as extracted from the special @page.author
	// variable
	Author string `json:"author,omitempty"`

	// list of categories the page belongs to, without the '.cat' extension
	Categories []string `json:"categories,omitempty"`

	// page title as extracted from the special @page.title variable, including
	// any possible HTML-encoded formatting
	FmtTitle wikifier.HTML `json:"fmt_title,omitempty"`

	// like FmtTitle except that all text formatting has been stripped.
	// suitable for use in the <title> tag
	Title string `json:"title,omitempty"`
}

// NewPage creates a Page given its filepath and configures it for
// use with this Wiki.
func (w *Wiki) NewPage(name string) *wikifier.Page {
	p := wikifier.NewPage(w.pathForPage(name, false, ""))
	p.Opt = &w.Opt
	return p
}

// DisplayPage returns the display result for a page.
func (w *Wiki) DisplayPage(name string) interface{} {
	return w.DisplayPageDraft(name, false)
}

// DisplayPageDraft returns the display result for a page.
//
// Unlike DisplayPage, if draftOK is true, the content is served even if it is
// marked as draft.
//
func (w *Wiki) DisplayPageDraft(name string, draftOK bool) interface{} {
	var r DisplayPage

	// create the page
	page := w.NewPage(name)

	// file does not exist
	if !page.Exists() {
		return DisplayError{
			Error:         "Page does not exist.",
			DetailedError: "Page '" + page.FilePath + "' does not exist.",
		}
	}

	// filename and path info
	path := page.Path()
	r.File = page.Name()
	r.Name = page.NameNE()
	r.Path = path

	// FIRST redirect check - symbolic link
	if page.IsSymlink() {
		return DisplayRedirect{Redirect: page.Redirect()}
	}

	// caching is enabled, so serve the cached copy if available
	if w.Opt.Page.EnableCache && page.CacheExists() {
		w.displayCachedPage(page, &r)
		return r
	}

	// Safe point - we will be generating the page right now.

	// parse the page
	//
	// if an error occurs, parse it again in variable-only mode.
	// then hopefully we can at least get the metadata and categories
	//
	err := page.Parse()
	if err != nil {
		page.VarsOnly = true
		page.Parse()
		// TODO: add page to categories
		// TODO: cache the error
		return DisplayError{Error: err.Error(), ParseError: true}
	}

	// if this is a draft and we're not serving drafts, pretend
	// that the page does not exist
	if !draftOK && page.Draft() {
		return DisplayError{Error: "Page has not yet been publised.", Draft: true}
	}

	// SECOND redirect check -
	// this is for pages we just parsed with @page.redirect
	if redir := page.Redirect(); redir != "" {
		return DisplayRedirect{Redirect: redir}
	}

	// generate HTML and metadata
	r.Generated = true
	r.Title = page.Title()
	r.FmtTitle = page.FmtTitle()
	r.Author = page.Author()
	r.Draft = page.Draft()
	r.ModUnix = page.Modified().Unix()
	r.Modified = httpdate.Time2Str(page.Modified())
	// r.Content = page.HTML()
	r.CSS = page.CSS()

	// TODO: should we include the page object?
	// TODO: warnings

	// TODO: update categories and set to r.Categories
	// TODO: write cache file if enabled
	// TODO: write search file if enabled

	return r
}

func (w *Wiki) displayCachedPage(page *wikifier.Page, r *DisplayPage) {
}
