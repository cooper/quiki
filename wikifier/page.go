package wikifier

import (
	"bufio"
	"bytes"
	"errors"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/cooper/quiki/markdown"
	strip "github.com/grokify/html-strip-tags-go"
)

// Page represents a single page or article, generally associated with a .page file.
// It provides the most basic public interface to parsing with the wikifier engine.
type Page struct {
	Source       string   // source content
	FilePath     string   // Path to the .page file
	VarsOnly     bool     // True if Parse() should only extract variables
	Opt          *PageOpt // page options
	styles       []styleEntry
	staticStyles []string
	codeStyles   bool
	parser       *parser              // wikifier parser instance
	main         block                // main block
	Images       map[string][][]int   // references to images
	Models       map[string]ModelInfo // references to models
	PageLinks    map[string][]int     // references to other pages
	sectionN     int
	name         string
	headingIDs   map[string]int
	Wiki         any       // only available during Parse() and HTML()
	Markdown     bool      // true if this is a markdown source
	model        bool      // true if this is a model being generated
	Warnings     []Warning // parser warnings
	Error        *Warning  // parser error, as an encodable Warning
	_html        HTML
	_text        string
	_preview     string
	_styleId     int
	*variableScope
}

// PageInfo represents metadata associated with a page.
type PageInfo struct {
	Path        string     `json:"-"`                   // absolute filepath
	File        string     `json:"file,omitempty"`      // name with extension, always with forward slashes
	FileNE      string     `json:"file_ne,omitempty"`   // name without extension, always with forward slashes
	Base        string     `json:"base,omitempty"`      // base name with extension
	BaseNE      string     `json:"base_ne,omitempty"`   // base name without extension
	Created     *time.Time `json:"created,omitempty"`   // creation time
	Modified    *time.Time `json:"modified,omitempty"`  // modify time
	Draft       bool       `json:"draft,omitempty"`     // true if page is marked as draft
	Generated   bool       `json:"generated,omitempty"` // true if page was generated from another source
	External    bool       `json:"external,omitempty"`  // true if page is outside the page directory
	Redirect    string     `json:"redirect,omitempty"`  // path page is to redirect to
	FmtTitle    HTML       `json:"fmt_title,omitempty"` // title with formatting tags
	Title       string     `json:"title,omitempty"`     // title without tags
	Author      string     `json:"author,omitempty"`    // author's name
	Description string     `json:"desc,omitempty"`      // description
	Keywords    []string   `json:"keywords,omitempty"`  // keywords
	Preview     string     `json:"preview,omitempty"`   // first 25 words or 150 chars. empty w/ description
	Warnings    []Warning  `json:"warnings,omitempty"`  // parser warnings
	Error       *Warning   `json:"error,omitempty"`     // parser error, as an encodable warning
}

// Warning represents a warning on a page.
type Warning struct {
	Message string   `json:"message"`
	Pos     Position `json:"position"`
}

// NewPage creates a page given its filepath.
func NewPage(filePath string) *Page {
	myOpt := defaultPageOpt // copy
	return &Page{
		FilePath:      filePath,
		Opt:           &myOpt,
		variableScope: newVariableScope(),
		Images:        make(map[string][][]int),
		Models:        make(map[string]ModelInfo),
		PageLinks:     make(map[string][]int),
		headingIDs:    make(map[string]int),
		Markdown:      strings.HasSuffix(filePath, ".md"),
	}
}

// NewPageSource creates a page given some source code.
func NewPageSource(source string) *Page {
	p := NewPage("")
	p.Source = source
	return p
}

// NewPagePath creates a page given its filepath and relative name.
func NewPagePath(filePath, name string) *Page {
	p := NewPage(filePath)
	p.name = name
	return p
}

// Parse opens the page file and attempts to parse it, returning any errors encountered.
func (p *Page) Parse() error {

	// create parser
	p.parser = newParser(p)
	p.main = p.parser.block
	defer p.resetParseState()

	// call underlying parse
	err := p._parse()
	if err == nil {
		return err
	}

	// error occurred--

	// if not already a ParserError
	var perr *ParserError
	if !errors.As(err, &perr) {
		// wrap to include current positional info
		perr = &ParserError{Pos: p.parser.pos, Err: err}
	}

	// convert to Warning for p.Error
	p.Error = &Warning{
		Message: perr.Err.Error(),
		Pos:     perr.Pos,
	}
	p.Error.Log(p.RelPath())

	return perr
}

func (p *Page) _parse() error {

	// create reader from file path or source code provided
	var reader io.Reader
	if p.Markdown && p.Source != "" {
		d := markdown.Run([]byte(p.Source))
		reader = bytes.NewReader(d)
	} else if p.Source != "" {
		reader = strings.NewReader(p.Source)
	} else if p.Markdown && p.FilePath != "" {
		// TODO: parse markdown as it's read instead of reading whole file to memory
		md, err := os.ReadFile(p.FilePath)
		if err != nil {
			return err
		}
		d := markdown.Run(md)
		reader = bytes.NewReader(d)
	} else if p.FilePath != "" {
		file, err := os.Open(p.FilePath)
		if err != nil {
			return err
		}
		defer file.Close()
		reader = file
	} else {
		return errors.New("neither Source nor FilePath provided")
	}

	// parse line-by-line
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := p.parser.parseLine(scanner.Bytes(), p); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// check if p.parser.catch != main block
	if p.parser.catch != p.main {
		blk := p.parser.block
		if p.parser.catch == blk {
			return parserError(blk.openPosition(), blk.blockType()+"{} not closed")
		}
		return errors.New(string(p.parser.catch.catchType()) + " not closed")
	}

	// parse the blocks, unless we only want vars
	if !p.VarsOnly {
		p.main.parse(p)
	}

	return nil
}

// HTML generates and returns the HTML code for the page.
// The page must be parsed with Parse before attempting this method.
func (p *Page) HTML() HTML {
	if p._html == "" {
		p._html = generateBlock(p.main, p)
	}
	return p._html
}

// HTMLAndCSS generates and returns the HTML code for the page, including CSS.
func (p *Page) HTMLAndCSS() HTML {
	css := p.CSS()
	if css != "" {
		return HTML("<style>\n" + css + "\n</style>\n" + string(p.HTML()))
	}
	return p.HTML()
}

// Text generates and returns the rendered plain text for the page.
// The page must be parsed with Parse before attempting this method.
func (p *Page) Text() string {
	if p._text != "" {
		return p._text
	}
	p._text = html.UnescapeString(strip.StripTags(string(p.HTML())))
	return p._text
}

// Preview returns a preview of the text on the page, up to 25 words or 150 characters.
// If the page has a Description, that is used instead of generating a preview.
func (p *Page) Preview() string {
	if p._preview != "" {
		return p._preview
	}

	text := p.Text()

	// remove excess whitespace
	preview := ""
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		preview += trimmed + " "
	}

	// find the first 25 words or 150 characters
	count := 25
	lastSpace := 0
	for i := range preview {

		// found space
		if preview[i] == ' ' {
			lastSpace = i
			count--

			// reached word limit
			if count == 0 {
				preview = preview[:lastSpace]
				break
			}
		}

		// reached character limit
		if i == 149 {
			preview = preview[:lastSpace]
			break
		}
	}

	preview = strings.TrimSpace(preview)
	p._preview = preview
	return preview
}

// Exists is true if the page exists.
func (p *Page) Exists() bool {
	if p.Source != "" {
		return true
	}
	_, err := os.Stat(p.FilePath)
	return err == nil
}

// CacheExists is true if the page cache file exists.
func (p *Page) CacheExists() bool {
	_, err := os.Stat(p.CachePath())
	return err == nil
}

// Name returns the resolved page name with extension.
//
// This DOES take symbolic links into account.
// and DOES include the page prefix if applicable.
// Any prefix will have forward slashes regardless of OS.
func (p *Page) Name() string {
	dir := pageAbs(p.Opt.Dir.Page)
	path := p.Path()

	// external; use relative
	if p.External() {
		return p.RelName()
	}

	if path == "" {
		return ""
	}

	// make relative to page directory
	if name, err := filepath.Rel(dir, path); err == nil {

		// remove possible leading slash
		name = strings.TrimPrefix(filepath.ToSlash(name), "/") // path/to/quiki/doc/language.md
		return name
	}

	// if the path cannot be made relative to the page dir,
	// it is probably a symlink to something external
	return p.RelName()
}

// OSName is like Name, except it uses the native path separator.
// It should be used for file operations only.
func (p *Page) OSName() string {
	return filepath.FromSlash(p.Name())
}

// NameNE returns the resolved page name with No Extension.
func (p *Page) NameNE() string {
	return PageNameNE(p.Name())
}

// OSNameNE is like NameNE, except it uses the native path separator.
// It should be used for file operations only.
func (p *Page) OSNameNE() string {
	return filepath.FromSlash(p.NameNE())
}

// Prefix returns the page prefix.
//
// For example, for a page named a/b.page, this is a.
// For a page named a.page, this is an empty string.
func (p *Page) Prefix() string {
	dir := strings.TrimSuffix(filepath.ToSlash(filepath.Dir(p.Name())), "/")
	if dir == "." {
		return ""
	}
	return dir
}

// Path returns the absolute path to the page as resolved.
// If the path does not resolve, returns an empty string.
func (p *Page) Path() string {
	relPath := p.RelPath()
	if relPath == "" {
		return ""
	}
	return pageAbs(relPath)
}

// RelName returns the unresolved page filename, with or without extension.
// This does NOT take symbolic links into account.
// It is not guaranteed to exist.
func (p *Page) RelName() string {

	// name is predetermined
	if p.name != "" {
		return p.name
	}

	dir := pageAbs(p.Opt.Dir.Page)
	path := p.RelPath() // this is what makes it different from Name()
	if path == "" {
		return ""
	}

	// make relative to page directory
	if name, err := filepath.Rel(dir, path); err == nil {

		// remove possible leading slash
		name = strings.TrimPrefix(filepath.ToSlash(name), "/") // path/to/quiki/doc/language.md
		return name
	}

	// if the path cannot be made relative to the page dir,
	// it is probably a symlink to something external
	return filepath.ToSlash(path)
}

// RelNameNE returns the unresolved page name with No Extension, relative to
// the page directory option.
// This does NOT take symbolic links into account.
// It is not guaranteed to exist.
func (p *Page) RelNameNE() string {
	return PageNameNE(p.RelName())
}

// RelPath returns the unresolved file path to the page.
// It may be a relative or absolute path.
// It is not guaranteed to exist.
func (p *Page) RelPath() string {
	if p.FilePath != "" {
		return p.FilePath
	}
	if p.name == "" {
		return ""
	}
	return filepath.Join(p.Opt.Dir.Page, p.name)
}

// getPageStr is like GetStr except it adds the proper prefix
// for @page and @model vars.
func (p *Page) getPageStr(key string) (string, error) {
	if p.model {
		return p.GetStr("model." + key)
	}
	return p.GetStr("page." + key)
}

// getPageBool is like GetBool except it adds the proper prefix
// for @page and @model vars.
func (p *Page) getPageBool(key string) (bool, error) {
	if p.model {
		return p.GetBool("model." + key)
	}
	return p.GetBool("page." + key)
}

// Redirect returns the location to which the page redirects, if any.
// This may be a relative or absolute URL, suitable for use in a Location header.
func (p *Page) Redirect() string {

	// symbolic link redirect
	if p.IsSymlink() {
		return pageAbs(filepath.Join(p.Opt.Root.Page, p.NameNE()))
	}

	// @page.redirect
	if link, err := p.getPageStr("redirect"); err != nil {
		// FIXME: is there anyway to produce a warning for wrong variable type?
	} else if ok, target, _, _, _ := parseLink(p.mainBlock(), link, &FmtOpt{}); ok {
		return target
	}

	return ""
}

// IsSymlink returns true if the page is a symbolic link to another file within
// the page directory. If it is symlinked to somewhere outside the page directory,
// it is treated as a normal page rather than a redirect.
func (p *Page) IsSymlink() bool {
	dirPage := pageAbs(p.Opt.Dir.Page)
	if !strings.HasPrefix(p.Path(), dirPage) {
		return false
	}
	fi, err := os.Lstat(p.RelPath())
	if err != nil || fi == nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// Created returns the page creation time.
func (p *Page) Created() time.Time {
	var t time.Time
	// FIXME: maybe produce a warning if this is not in the right format
	created, _ := p.getPageStr("created")
	if created == "" {
		return t
	}
	if unix, err := strconv.ParseInt(created, 10, 0); err == nil {
		return time.Unix(unix, 0)
	}
	t, _ = httpdate.Str2Time(created, time.UTC)
	return t
}

// Modified returns the page modification time.
func (p *Page) Modified() time.Time {
	fi, err := os.Lstat(p.Path())
	if err != nil || fi == nil {
		return time.Time{}
	}
	return fi.ModTime()
}

// CachePath returns the absolute path to the page cache file.
func (p *Page) CachePath() string {
	osName := p.OSName() + ".cache" // os-specific cache name
	MakeDir(filepath.Join(p.Opt.Dir.Cache, "page"), osName)
	return pageAbs(filepath.Join(p.Opt.Dir.Cache, "page", osName))
}

// CacheModified returns the page cache file time.
func (p *Page) CacheModified() time.Time {
	fi, err := os.Lstat(p.CachePath())
	if err != nil || fi == nil {
		return time.Time{} // return zero time if cache file doesn't exist
	}
	return fi.ModTime()
}

// SearchPath returns the absolute path to the page search text file.
func (p *Page) SearchPath() string {
	osName := p.OSName() + ".txt" // os-specific text file name
	MakeDir(filepath.Join(p.Opt.Dir.Cache, "page"), osName)
	return pageAbs(filepath.Join(p.Opt.Dir.Cache, "page", osName))
}

// Draft returns true if the page is marked as a draft.
func (p *Page) Draft() bool {
	b, _ := p.getPageBool("draft")
	return b
}

// Generated returns true if the page was auto-generated
// from some other source content.
func (p *Page) Generated() bool {
	b, _ := p.getPageBool("generated")
	return b
}

// External returns true if the page is outside the page directory
// as defined by the configuration, with symlinks considered.
//
// If `dir.wiki` isn't set, External is always true
// (since the page is not associated with a wiki at all).
func (p *Page) External() bool {

	// not part of a wiki at all
	dirPage := pageAbs(p.Opt.Dir.Page)
	if dirPage == "" {
		return true
	}

	// cannot be made relative
	rel, err := filepath.Rel(dirPage, p.Path())
	if err != nil {
		return true
	}

	// contains ../ so it's not relative
	if strings.Contains(rel, ".."+string(os.PathSeparator)) {
		return true
	}
	if strings.Contains(rel, string(os.PathSeparator)+"..") {
		return true
	}

	// otherwise it's in there
	return false
}

// Author returns the page author's name, if any.
func (p *Page) Author() string {
	s, _ := p.getPageStr("author")
	return s
}

// FmtTitle returns the page title, preserving any possible text formatting.
func (p *Page) FmtTitle() HTML {
	s, _ := p.getPageStr("title")
	return HTML(s)
}

// Title returns the page title with HTML text formatting tags stripped.
func (p *Page) Title() string {
	return strip.StripTags(string(p.FmtTitle()))
}

// TitleOrName returns the result of Title if available, otherwise that of Name.
func (p *Page) TitleOrName() string {
	if title := p.Title(); title != "" {
		return title
	}
	return p.Name()
}

// Description returns the page description.
func (p *Page) Description() string {
	s, _ := p.getPageStr("desc")
	if s == "" {
		s, _ = p.getPageStr("description")
	}
	return strip.StripTags(html.UnescapeString(s))
}

// Keywords returns the list of page keywords.
func (p *Page) Keywords() []string {
	list, _ := p.GetStrList("page.keywords")
	return list
}

// Categories returns a list of categories the page belongs to.
func (p *Page) Categories() []string {
	obj, err := p.GetObj("category")
	if err != nil {
		return nil
	}
	catMap, ok := obj.(*Map)
	if !ok {
		return nil
	}
	return catMap.Keys()
}

// Info returns the PageInfo for the page.
func (p *Page) Info() PageInfo {

	// generate preview only if there is no description available
	prev := ""
	desc := p.Description()
	if desc == "" {
		prev = p.Preview()
	}

	// info
	info := PageInfo{
		File:        p.Name(),
		FileNE:      p.NameNE(),
		Base:        filepath.Base(p.Name()),
		BaseNE:      filepath.Base(p.NameNE()),
		Draft:       p.Draft(),
		Generated:   p.Generated(),
		External:    p.External(),
		Redirect:    p.Redirect(),
		FmtTitle:    p.FmtTitle(),
		Title:       p.Title(),
		Author:      p.Author(),
		Description: desc,
		Keywords:    p.Keywords(),
		Preview:     prev,
		Warnings:    p.Warnings,
		Error:       p.Error,
	}

	// file times
	mod, create := p.Modified(), p.Created()
	if !mod.IsZero() {
		info.Modified = &mod
		info.Created = &mod // fallback
	}
	if !create.IsZero() {
		info.Created = &create
	}

	return info
}

// create a page warning
func (p *Page) warn(pos Position, warning string) {
	w := Warning{warning, pos}
	p.Warnings = append(p.Warnings, w)
	w.Log(p.RelPath())
}

func (p *Page) mainBlock() block {
	return p.main
}

// resets the parser
func (p *Page) resetParseState() {
	p.parser = nil
}

func pageAbs(path string) string {
	if abs, _ := filepath.Abs(path); abs != "" {
		path = abs
	}
	if followed, _ := filepath.EvalSymlinks(path); followed != "" {
		return followed
	}
	return path
}

func (w Warning) Log(path string) {
	log.Printf("%s:%d:%d: %s", path, w.Pos.Line, w.Pos.Column, w.Message)
}
