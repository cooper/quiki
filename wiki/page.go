package wiki

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	Content wikifier.HTML `json:"-"`

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

	// warnings and errors produced by the parser
	Warnings []wikifier.Warning `json:"warnings,omitempty"`
	Errors   []wikifier.Warning `json:"errors,omitempty"`

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

	// page description as extracted from the special @page.desc variable.
	Description string `json:"desc,omitempty"`

	// page keywords as extracted from the special @page.keywords variable
	Keywords []string `json:"keywords,omitempty"`
}

type pageJSONManifest struct {
	CSS        string   `json:"css,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Error      string   `json:"error,omitempty"`
	wikifier.PageInfo
}

// FindPage attempts to find a page on this wiki given its name,
// regardless of the file format or filename case.
//
// If a page by this name exists, the returned page represents it.
// Otherwise, a new page representing the lowercased, normalized .page
// file is returned in the standard quiki filename format.
//
func (w *Wiki) FindPage(name string) (p *wikifier.Page) {

	// separate into prefix and base
	pfx, base := filepath.Dir(name), filepath.Base(name)

	// try in this order
	tryFiles := []string{
		wikifier.PageNameLink(base),                            // exact match with no lowercasing
		wikifier.PageNameLink(base) + ".page",                  // .page with no lowercasing
		strings.ToLower(wikifier.PageNameLink(base)) + ".page", // .page with lowercasing
		wikifier.PageNameLink(base) + ".md",                    // .md with no lowercasing
		strings.ToLower(wikifier.PageNameLink(base)) + ".md",   // .md with lowercasing
	}
	path := ""
	for _, try := range tryFiles {
		path = filepath.Join(w.Opt.Dir.Page, pfx, try)
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			break
		}
		path = ""
	}

	// found a match!
	if path != "" {
		p = wikifier.NewPagePath(path, name) // consider: name case might be wrong?
	} else {

		// didn't find anything, so create one
		p = wikifier.NewPagePath(w.pathForPage(wikifier.PageName(name)), name)
	}

	// these are available to all pages
	p.Wiki = w
	p.Opt = &w.Opt

	// create page lock
	if _, exist := w.pageLocks[p.Name()]; !exist {
		w.pageLocks[p.Name()] = new(sync.Mutex)
	}

	return
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
	page := w.FindPage(name)

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
		var pErr *wikifier.ParserError
		errors.As(err, &pErr)
		return DisplayError{Error: err.Error(), ParseError: pErr}
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
		w.updatePageCategories(page)
		// consider: set r.Categories? can redirects belong to categories?
		return DisplayRedirect{Redirect: redir}
	}

	// only generate once at a time
	w.pageLocks[r.File].Lock()
	defer w.pageLocks[r.File].Unlock()

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
	r.Description = page.Description()
	r.Keywords = page.Keywords()
	r.Draft = page.Draft()
	r.Modified = &mod
	r.ModifiedHTTP = httpdate.Time2Str(mod)
	r.Content = page.HTML()
	r.CSS = page.CSS()
	r.Warnings = page.Warnings

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

// Pages returns info about all the pages in the wiki.
func (w *Wiki) Pages() []wikifier.PageInfo {
	pageNames := w.allPageFiles()
	pages := make([]wikifier.PageInfo, len(pageNames))

	// pages individually
	i := 0
	for _, name := range pageNames {
		pages[i] = w.PageInfo(name)
		i++
	}

	return pages
}

type sortablePageInfo wikifier.PageInfo

func (pi sortablePageInfo) SortInfo() SortInfo {
	return SortInfo{
		Title:    pi.Title,
		Author:   pi.Author,
		Created:  *pi.Created,
		Modified: *pi.Modified,
	}
}

// PagesSorted returns info about all the pages in the wiki, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, and SortModified.
func (w *Wiki) PagesSorted(descend bool, sorters ...SortFunc) []wikifier.PageInfo {

	// convert to []Sortable
	pages := w.Pages()
	sorted := make([]Sortable, len(pages))
	for i, pi := range w.Pages() {
		sorted[i] = sortablePageInfo(pi)
	}

	// sort
	var sorter sort.Interface = sorter(sorted, sorters...)
	if descend {
		sorter = sort.Reverse(sorter)
	}
	sort.Sort(sorter)

	// convert back to []wikifier.PageInfo
	for i, si := range sorted {
		pages[i] = wikifier.PageInfo(si.(sortablePageInfo))
	}

	return pages
}

// PageMap returns a map of page name to PageInfo for all pages in the wiki.
func (w *Wiki) PageMap() map[string]wikifier.PageInfo {
	pageNames := w.allPageFiles()
	pages := make(map[string]wikifier.PageInfo, len(pageNames))

	// pages individually
	for _, name := range pageNames {
		pages[name] = w.PageInfo(name)
	}

	return pages
}

// PageInfo is an inexpensive request for info on a page. It uses cached
// metadata rather than generating the page and extracting variables.
func (w *Wiki) PageInfo(name string) (info wikifier.PageInfo) {

	// the page does not exist
	path := w.pathForPage(name)
	pgFi, err := os.Stat(path)
	if err != nil {
		return
	}

	// find page category
	nameNE := wikifier.PageNameNE(name)
	pageCat := w.GetSpecialCategory(nameNE, CategoryTypePage)

	// if page category exists use that info
	if pageCat.Exists() && pageCat.PageInfo != nil {
		info = *pageCat.PageInfo
	}

	// this stuff is available to all
	mod := pgFi.ModTime()
	info.Path = path
	info.File = filepath.ToSlash(name)
	info.Modified = &mod // actual page mod time

	// fallback title to name
	if info.Title == "" {
		info.Title = nameNE
	}

	// fallback created to modified
	if info.Created == nil {
		info.Created = &mod
	}

	return
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
	cacheFile.Write([]byte{'\n'})
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
		Error:      "", // TODO
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
			DetailedError: "JSON decode error: " + err.Error(),
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
	r.Description = info.Description
	r.Keywords = info.Keywords
	r.Warnings = info.Warnings
	r.FromCache = true
	r.CSS = info.CSS
	r.Content = wikifier.HTML(content)
	r.Modified = &cacheModify
	r.ModifiedHTTP = httpdate.Time2Str(cacheModify)

	return nil // success
}
