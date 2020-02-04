package wiki

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	strip "github.com/grokify/html-strip-tags-go"

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

	// time when the page was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the page file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted for Last-Modified

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

	// time when the page was created, as extracted from
	// the special @page.created variable
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

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

type pageJSONManifest struct {
	CSS        string   `json:"css,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Error      string   `json:"error,omitempty"`
	wikifier.PageInfo
}

// NewPage creates a Page given its name and configures it for
// use with this Wiki.
func (w *Wiki) NewPage(name string) *wikifier.Page {

	// lowercase .page exists
	p := w._newPage(name)
	if p.Exists() {
		return p
	}

	// lowercase .md exists
	if mdp := w._newPage(name + ".md"); mdp.Exists() {
		return mdp
	}

	return p
}

func (w *Wiki) _newPage(name string) *wikifier.Page {
	p := wikifier.NewPageNamed(w.pathForPage(name, false, ""), name)
	p.Wiki = w
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
		if errOrRedir := w.displayCachedPage(page, &r, draftOK); errOrRedir != nil {
			return errOrRedir
		}
		if r.FromCache {
			return r
		}
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

	// THIRD redirect check -
	// this is for pages we just parsed with @page.redirect
	if redir := page.Redirect(); redir != "" {
		w.writeRedirectCache(page)
		return DisplayRedirect{Redirect: redir}
	}

	// generate HTML and metadata
	create := page.Created()
	if !create.IsZero() {
		r.Created = &create
		r.CreatedHTTP = httpdate.Time2Str(create)
	}
	mod := page.Modified()
	r.Generated = true
	r.Title = page.Title()
	r.FmtTitle = page.FmtTitle()
	r.Author = page.Author()
	r.Draft = page.Draft()
	r.Modified = &mod
	r.ModifiedHTTP = httpdate.Time2Str(mod)
	r.Content = page.HTML()
	r.CSS = page.CSS()

	// TODO: should we include the page object?
	// TODO: warnings

	// update categories
	w.updatePageCategories(page)
	r.Categories = page.Categories()

	// only do these things if content was generated
	if !page.VarsOnly {

		// write cache file if enabled
		if dispErr := w.writePageCache(page, &r); dispErr != nil {
			return dispErr
		}

		// write search file if enabled
		if dispErr := w.writePageText(page, &r); dispErr != nil {
			return dispErr
		}
	}

	return r
}

func (w *Wiki) writeRedirectCache(page *wikifier.Page) {

	// caching isn't enabled
	if !page.Opt.Page.EnableCache || page.CachePath() == "" {
		return
	}

	// open the cache file for writing
	cacheFile, err := os.Create(page.CachePath())
	defer cacheFile.Close()
	if err != nil {
		return
	}

	// create manifest with just page info (includes redirect)
	j, err := json.Marshal(pageJSONManifest{PageInfo: page.Info()})
	if err != nil {
		return
	}

	cacheFile.Write(j)
}

func (w *Wiki) writePageCache(page *wikifier.Page, r *DisplayPage) interface{} {

	// caching isn't enabled
	if !page.Opt.Page.EnableCache || page.CachePath() == "" {
		return nil
	}

	// open the cache file for writing
	cacheFile, err := os.Create(page.CachePath())
	defer cacheFile.Close()
	if err != nil {
		return DisplayError{
			Error:         "Could not write page cache file.",
			DetailedError: "Open '" + page.CachePath() + "' for write error: " + err.Error(),
		}
	}

	// generate page info
	info := pageJSONManifest{
		CSS:        r.CSS,
		Categories: r.Categories,
		Warnings:   []string{}, // TODO
		Error:      "",         // TODO
		PageInfo:   page.Info(),
	}

	// encode as json
	j, err := json.Marshal(info)
	if err != nil {
		return DisplayError{
			Error:         "Could not write page cache file.",
			DetailedError: "JSON encode error: " + err.Error(),
		}
	}

	// save prefixing data
	cacheFile.Write(j)
	cacheFile.Write([]byte{'\n'})

	// save content
	content := string(r.Content)
	cacheFile.WriteString(content)
	if len(content) != 0 && content[len(content)-1] != '\n' {
		cacheFile.Write([]byte{'\n'})
	}

	// update result with real cache modified times
	mod := page.CacheModified()
	r.Modified = &mod
	r.ModifiedHTTP = httpdate.Time2Str(mod)
	r.CacheGenerated = true

	return nil // success
}

func (w *Wiki) writePageText(page *wikifier.Page, r *DisplayPage) interface{} {

	// search optimization isn't enabled
	if !page.Opt.Search.Enable || page.SearchPath() == "" || r.Content == "" {
		return nil
	}

	// open the text file for writing
	textFile, err := os.Create(page.SearchPath())
	defer textFile.Close()
	if err != nil {
		return DisplayError{
			Error:         "Could not write page text file.",
			DetailedError: "Open '" + page.SearchPath() + "' for write error: " + err.Error(),
		}
	}

	// save the content with HTML tags stripped
	textFile.WriteString(strip.StripTags(string(r.Content)))

	r.TextGenerated = true
	return nil // success
}

func (w *Wiki) displayCachedPage(page *wikifier.Page, r *DisplayPage, draftOK bool) interface{} {
	cacheModify := page.CacheModified()
	timeStr := httpdate.Time2Str(cacheModify)

	// the page's file is more recent than the cache file.
	// discard the outdated cached copy
	if page.Modified().After(cacheModify) {
		os.Remove(page.CachePath())
		return nil // OK
	}

	content := "<!-- cached page dated " + timeStr + " -->\n"

	// open cache file for reading
	cacheContent, err := ioutil.ReadFile(page.CachePath())
	if err != nil {
		return DisplayError{
			Error:         "Could not read page cache file.",
			DetailedError: "Open '" + page.CachePath() + "' for read error: " + err.Error(),
		}
	}

	// find the first line
	var jsonData []byte
	firstNL := bytes.IndexByte(cacheContent, '\n')
	if firstNL != -1 && firstNL < len(cacheContent) {
		jsonData = cacheContent[:firstNL+1]
		cacheContent = cacheContent[firstNL:]
	}

	// the rest is the html content
	content += string(cacheContent)

	// decode the manifest
	var info pageJSONManifest
	if err := json.Unmarshal(jsonData, &info); err != nil {
		return DisplayError{
			Error:         "Could not read page cache file.",
			DetailedError: "JSON encode error: " + err.Error(),
		}
	}

	// if this is a draft and we're not serving drafts, pretend
	// that the page does not exist
	if !draftOK && info.Draft {
		return DisplayError{Error: "Page has not yet been publised.", Draft: true}
	}

	// cached error
	if info.Error != "" {
		return DisplayError{Error: info.Error}
	}

	// cached redirect
	if info.Redirect != "" {
		return DisplayRedirect{Redirect: info.Redirect}
	}

	// update result with stuff from cache
	if info.Created != nil {
		r.Created = info.Created
		r.CreatedHTTP = httpdate.Time2Str(*info.Created)
	}
	r.Draft = info.Draft
	r.Author = info.Author
	r.Title = info.Title
	r.FmtTitle = info.FmtTitle
	r.FromCache = true
	r.CSS = info.CSS
	r.Content = wikifier.HTML(content)
	r.Modified = &cacheModify
	r.ModifiedHTTP = httpdate.Time2Str(cacheModify)

	return nil // success
}
