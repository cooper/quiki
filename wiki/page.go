package wiki

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/cooper/quiki/adminifier/utils"
	"github.com/cooper/quiki/wikifier"
)

// DisplayPage represents a page result to display.
type DisplayPage struct {

	// name of the page relative to the pages dir, with the extension; e.g. some/page.page
	File string `json:"file,omitempty"`

	// name of the page without the extension; e.g. some/page
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

	// first formatting-stripped 25 words of page, up to 150 chars
	Preview string `json:"preview,omitempty"`
}

type pageJSONManifest struct {
	CSS        string   `json:"css,omitempty"`
	Categories []string `json:"categories,omitempty"`
	wikifier.PageInfo
}

// FindPage attempts to find a page on this wiki given its name,
// regardless of the file format or filename case.
//
// If a page by this name exists, the returned page represents it.
// Otherwise, a new page representing the lowercased, normalized .page
// file is returned in the standard quiki filename format.
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
		p = wikifier.NewPagePath(w.PathForPage(wikifier.PageName(name)), name)
	}

	// these are available to all pages
	p.Wiki = w
	// give each page its own copy of the wiki options
	// so that page-specific options don't affect other pages
	pageOpt := w.Opt // copy the struct
	p.Opt = &pageOpt

	// page lock will be created by GetPageLock if needed
	return
}

// DisplayPage returns the display result for a page.
func (w *Wiki) DisplayPage(name string) any {
	return w.DisplayPageDraft(name, false)
}

// DisplayPageDraft returns the display result for a page.
//
// Unlike DisplayPage, if draftOK is true, the content is served even if it is
// marked as draft.
func (w *Wiki) DisplayPageDraft(name string, draftOK bool) any {
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
	if w.Opt.Page.EnableCache && !w.Opt.Page.ForceGen && page.CacheExists() {
		if errOrRedir := w.displayCachedPage(page, &r, draftOK); errOrRedir != nil {
			return errOrRedir
		}
		if r.FromCache {
			return r
		}
	} else if w.Opt.Page.EnableCache {
		// Cache doesn't exist - page will be generated
		// Note: Pregeneration is now handled at the webserver level
	}

	// Safe point - we will be generating the page right now.

	// parse the page
	//
	// if an error occurs, parse it again in variable-only mode.
	// then hopefully we can at least get the metadata and categories
	//
	err := page.Parse()
	if err != nil {

		// copy original error because we will reparse
		oldErr := page.Error

		// re-parse in variable-only mode to extract page data
		page.VarsOnly = true
		page.Parse()

		// write original error to cache
		page.Error = oldErr
		w.writeVarsCache(page)

		// add page to categories-
		// should be possible if VarsOnly mode was successful
		w.updatePageCategories(page)

		// extract the ParserError
		var pErr *wikifier.ParserError
		errors.As(err, &pErr)

		return DisplayError{Error: err.Error(), Pos: pErr.Pos}
	}

	// if this is a draft and we're not serving drafts, pretend
	// that the page does not exist
	if !draftOK && page.Draft() {
		return DisplayError{Error: "Page has not yet been published.", Draft: true}
	}

	// THIRD redirect check -
	// this is for pages we just parsed with @page.redirect
	if redir := page.Redirect(); redir != "" {
		w.writeVarsCache(page)
		w.updatePageCategories(page)
		// consider: set r.Categories? can redirects belong to categories?
		return DisplayRedirect{Redirect: redir}
	}

	// only generate once at a time
	pageLock := w.GetPageLock(r.File)
	pageLock.Lock()
	defer pageLock.Unlock()

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
	return w.pagesIn("", pageNames)
}

// PagesInDir returns info about all the pages in the specified directory.
func (w *Wiki) PagesInDir(where string) []wikifier.PageInfo {
	pageNames := w.pageFilesInDir(where)
	return w.pagesIn(where, pageNames)
}

func (w *Wiki) pagesIn(prefix string, pageNames []string) []wikifier.PageInfo {
	pages := make([]wikifier.PageInfo, len(pageNames))

	// pages individually
	i := 0
	for _, name := range pageNames {
		pages[i] = w.PageInfo(filepath.Join(prefix, name))
		i++
	}

	return pages
}

type sortablePageInfo wikifier.PageInfo

func (pi sortablePageInfo) SortInfo() SortInfo {
	title := pi.Title
	if title == "" {
		title = pi.BaseNE
	}
	return SortInfo{
		Title:    title,
		Author:   pi.Author,
		Created:  *pi.Created,
		Modified: *pi.Modified,
	}
}

// PagesSorted returns info about all the pages in the wiki, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, and SortModified.
func (w *Wiki) PagesSorted(descend bool, sorters ...SortFunc) []wikifier.PageInfo {
	return _pagesSorted(w.Pages(), descend, sorters...)
}

func _pagesSorted(pages []wikifier.PageInfo, descend bool, sorters ...SortFunc) []wikifier.PageInfo {
	// convert to []Sortable
	sorted := make([]Sortable, len(pages))
	for i, pi := range pages {
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

// PagesAndDirs returns info about all the pages and directories in a directory.
func (w *Wiki) PagesAndDirs(where string) ([]wikifier.PageInfo, []string) {
	pages := w.PagesInDir(where)
	dirs := utils.DirsInDir(filepath.Join(w.Opt.Dir.Page, where))
	return pages, dirs
}

// PagesAndDirsSorted returns info about all the pages and directories in a directory, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, and SortModified.
// Directories are always sorted alphabetically (but still respect the descend flag).
func (w *Wiki) PagesAndDirsSorted(where string, descend bool, sorters ...SortFunc) ([]wikifier.PageInfo, []string) {
	pages, dirs := w.PagesAndDirs(where)
	pages = _pagesSorted(pages, descend, sorters...)
	if descend {
		sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	} else {
		sort.Strings(dirs)
	}
	return pages, dirs
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
	path := w.PathForPage(name)
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

	// add info we can derive from the file
	mod := pgFi.ModTime()
	info.Path = path
	info.File = filepath.ToSlash(name)
	info.FileNE = filepath.ToSlash(nameNE)
	info.Base = filepath.Base(name)
	info.BaseNE = filepath.Base(nameNE)
	info.Modified = &mod // actual page mod time

	// fallback created to modified
	if info.Created == nil {
		info.Created = &mod
	}

	return
}

// like writePageCache except it only includes PageInfo.
// used for redirects and parser errors where vars could still be extracted.
func (w *Wiki) writeVarsCache(page *wikifier.Page) {

	// caching isn't enabled
	if !page.Opt.Page.EnableCache || page.CachePath() == "" {
		return
	}

	// open the cache file for writing
	cacheFile, err := os.Create(page.CachePath())
	if err != nil {
		return
	}
	defer cacheFile.Close()

	// create manifest with just page info (includes redirect/error)
	j, err := json.Marshal(pageJSONManifest{PageInfo: page.Info()})
	if err != nil {
		return
	}

	cacheFile.Write(j)
	cacheFile.Write([]byte{'\n'})
}

func (w *Wiki) writePageCache(page *wikifier.Page, r *DisplayPage) any {

	// caching isn't enabled
	if !page.Opt.Page.EnableCache || page.CachePath() == "" {
		return nil
	}

	// open the cache file for writing
	cacheFile, err := os.Create(page.CachePath())
	if err != nil {
		return DisplayError{
			Error:         "Could not write page cache file.",
			DetailedError: "Open '" + page.CachePath() + "' for write error: " + err.Error(),
		}
	}
	defer cacheFile.Close()

	// generate page info
	info := pageJSONManifest{
		CSS:        r.CSS,
		Categories: r.Categories,
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

func (w *Wiki) writePageText(page *wikifier.Page, r *DisplayPage) any {

	// search optimization isn't enabled
	if !page.Opt.Search.Enable || page.SearchPath() == "" || r.Content == "" {
		return nil
	}

	// open the text file for writing
	textFile, err := os.Create(page.SearchPath())
	if err != nil {
		return DisplayError{
			Error:         "Could not write page text file.",
			DetailedError: "Open '" + page.SearchPath() + "' for write error: " + err.Error(),
		}
	}
	defer textFile.Close()

	// save the content with HTML tags stripped
	textFile.WriteString(page.Text())

	r.TextGenerated = true
	return nil // success
}

func (w *Wiki) displayCachedPage(page *wikifier.Page, r *DisplayPage, draftOK bool) any {
	cacheModify := page.CacheModified()
	pageModified := page.Modified()
	timeStr := httpdate.Time2Str(cacheModify)

	// the page's file is more recent than the cache file.
	// discard the outdated cached copy
	if pageModified.After(cacheModify) {
		os.Remove(page.CachePath())
		return nil // OK
	}

	content := "<!-- cached page dated " + timeStr + " -->\n"

	// open cache file for reading
	cacheContent, err := os.ReadFile(page.CachePath())
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
		return DisplayError{
			Error: "Page has not yet been published.",
			Draft: true,
		}
	}

	// cached error
	if info.Error != nil {
		return DisplayError{
			Error: info.Error.Message,
			Pos:   info.Error.Pos,
		}
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
	r.Modified = &pageModified
	r.ModifiedHTTP = httpdate.Time2Str(pageModified)

	return nil // success
}

// like page.warn
func pageWarn(p *wikifier.Page, warning string, pos wikifier.Position) {
	w := wikifier.Warning{Message: warning, Pos: pos}
	p.Warnings = append(p.Warnings, w)
}

// RegeneratePage clears the cache for a page to force regeneration
func (w *Wiki) RegeneratePage(pageName string) error {
	page := w.FindPage(pageName)
	if !page.Exists() {
		return nil // Nothing to regenerate
	}

	// Clear the cache if it exists
	if page.CacheExists() {
		cachePath := page.CachePath()
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		w.Log(fmt.Sprintf("cleared cache for page: %s", pageName))
	}

	return nil
}
