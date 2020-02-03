package wikifier

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	Source     string   // source content
	FilePath   string   // Path to the .page file
	VarsOnly   bool     // True if Parse() should only extract variables
	Opt        *PageOpt // page options
	styles     []styleEntry
	parser     *parser            // wikifier parser instance
	main       block              // main block
	Images     map[string][][]int // references to images
	Models     map[string]bool    // references to models
	PageLinks  map[string][]int   // references to other pages
	sectionN   int
	name       string
	headingIDs map[string]int
	Wiki       interface{} // only available during Parse() and HTML()
	markdown   bool        // true if FilePath points to a markdown source
	*variableScope
}

// PageInfo represents metadata associated with a page.
type PageInfo struct {
	Created   *time.Time `json:"created,omitempty"`   // creation time
	Modified  *time.Time `json:"modified,omitempty"`  // modify time
	Draft     bool       `json:"draft,omitempty"`     // true if page is marked as draft
	Generated bool       `json:"generated,omitempty"` // true if page was generated from another source
	Redirect  string     `json:"redirect,omitempty"`  // path page is to redirect to
	FmtTitle  HTML       `json:"fmt_title,omitempty"` // title with formatting tags
	Title     string     `json:"title,omitempty"`     // title without tags
	Author    string     `json:"author,omitempty"`    // author's name
}

// NewPage creates a page given its filepath.
func NewPage(filePath string) *Page {
	myOpt := defaultPageOpt // copy
	return &Page{
		FilePath:      filePath,
		Opt:           &myOpt,
		variableScope: newVariableScope(),
		Images:        make(map[string][][]int),
		Models:        make(map[string]bool),
		PageLinks:     make(map[string][]int),
		headingIDs:    make(map[string]int),
		markdown:      strings.HasSuffix(filePath, ".md"),
	}
}

// NewPageSource creates a page given some source code.
func NewPageSource(source string) *Page {
	p := NewPage("")
	p.Source = source
	return p
}

// NewPageNamed creates a page given its filepath and relative name.
func NewPageNamed(filePath, name string) *Page {
	p := NewPage(filePath)
	p.name = name
	return p
}

// Parse opens the page file and attempts to parse it, returning any errors encountered.
func (p *Page) Parse() error {
	p.parser = newParser()
	p.main = p.parser.block
	defer p.resetParseState()

	// create reader from file path or source code provided
	var reader io.Reader
	if p.Source != "" {
		reader = strings.NewReader(p.Source)
	} else if p.markdown && p.FilePath != "" {
		md, err := ioutil.ReadFile(p.FilePath)
		if err != nil {
			return err
		}
		d := markdown.Run(md)
		fmt.Println(string(d))
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

	// TODO: check if p.parser.catch != main block
	if p.parser.catch != p.main {
		if p.parser.catch == p.parser.block {
			return fmt.Errorf("%s{} not closed; started at %d", p.parser.block.blockType(), p.parser.block.openPosition())
		}
		return errors.New(string(p.parser.catch.catchType()) + " not closed")
	}

	// parse the blocks, unless we only want vars
	if !p.VarsOnly {
		p.main.parse(p)
	}

	// inject variables set in the page to page opts
	if err := InjectPageOpt(p, p.Opt); err != nil {
		// TODO: position
		return err
	}

	return nil
}

// HTML generates and returns the HTML code for the page.
// The page must be parsed with Parse before attempting this method.
func (p *Page) HTML() HTML {
	// TODO: cache and then recursively destroy elements
	return generateBlock(p.main, p)
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

// Name returns the resolved page name, with or without extension.
//
// This DOES take symbolic links into account
// and DOES include the page prefix if applicable.
//
func (p *Page) Name() string {
	dir := pageAbs(p.Opt.Dir.Page)        // /path/to/quiki/wikis/mywiki/pages
	path := filepath.ToSlash(p.Path())    // /path/to/quiki/doc/language.md
	name := strings.TrimPrefix(path, dir) // /path/to/quiki/doc/language.md
	name = strings.TrimPrefix(name, "/")  // path/to/quiki/doc/language.md
	if strings.Index(path, dir) != 0 {    // if path does not start with /path/to/quiki/wikis/mywiki/pages
		return p.RelName() // return language.md
	}
	return name
}

// NameNE returns the resolved page name with No Extension.
func (p *Page) NameNE() string {
	return PageNameNE(p.Name())
}

// Prefix returns the page prefix.
//
// For example, for a page named a/b.page, this is a.
// For a page named a.page, this is an empty string.
//
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
	return pageAbs(p.RelPath())
}

// RelName returns the unresolved page filename, with or without extension.
// This does NOT take symbolic links into account.
// It is not guaranteed to exist.
func (p *Page) RelName() string {
	if p.name != "" {
		return p.name
	}
	dir := pageAbs(p.Opt.Dir.Page) // /path/to/quiki/wikis/mywiki/pages
	path := p.RelPath()            // doc/parsing.md
	name := strings.TrimPrefix(path, dir)
	name = strings.TrimPrefix(filepath.ToSlash(name), "/")
	if strings.Index(path, dir) == 0 {
		return filepath.Base(p.RelPath())
	}
	return name
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
	return filepath.Join(p.Opt.Dir.Page, p.name)
}

// Redirect returns the location to which the page redirects, if any.
// This may be a relative or absolute URL, suitable for use in a Location header.
func (p *Page) Redirect() string {

	// symbolic link redirect
	if p.IsSymlink() {
		return pageAbs(filepath.Join(p.Opt.Root.Page, p.NameNE()))
	}

	// @page.redirect
	if link, err := p.GetStr("page.redirect"); err != nil {
		// FIXME: is there anyway to produce a warning for wrong variable type?
	} else if ok, target, _, _, _ := p.parseLink(link); ok {
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
	fi, _ := os.Lstat(p.RelPath())
	return fi.Mode()&os.ModeSymlink != 0
}

// Created returns the page creation time.
func (p *Page) Created() time.Time {
	var t time.Time
	// FIXME: maybe produce a warning if this is not in the right format
	created, _ := p.GetStr("page.created")
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
	fi, _ := os.Lstat(p.Path())
	return fi.ModTime()
}

// CachePath returns the absolute path to the page cache file.
func (p *Page) CachePath() string {
	osName := filepath.FromSlash(p.Name()) + ".cache" // os-specific cache name
	MakeDir(filepath.Join(p.Opt.Dir.Cache, "page"), osName)
	return pageAbs(filepath.Join(p.Opt.Dir.Cache, "page", osName))
}

// CacheModified returns the page cache file time.
func (p *Page) CacheModified() time.Time {
	fi, _ := os.Lstat(p.CachePath())
	return fi.ModTime()
}

// SearchPath returns the absolute path to the page search text file.
func (p *Page) SearchPath() string {
	osName := filepath.FromSlash(p.Name()) + ".txt" // os-specific text file name
	MakeDir(filepath.Join(p.Opt.Dir.Cache, "page"), osName)
	return pageAbs(filepath.Join(p.Opt.Dir.Cache, "page", osName))
}

// Draft returns true if the page is marked as a draft.
func (p *Page) Draft() bool {
	b, _ := p.GetBool("page.draft")
	return b
}

// Generated returns true if the page was auto-generated
// from some other source content.
func (p *Page) Generated() bool {
	b, _ := p.GetBool("page.generated")
	return b
}

// Author returns the page author's name, if any.
func (p *Page) Author() string {
	s, _ := p.GetStr("page.author")
	return s
}

// FmtTitle returns the page title, preserving any possible text formatting.
func (p *Page) FmtTitle() HTML {
	s, _ := p.GetStr("page.title")
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
	info := PageInfo{
		Draft:     p.Draft(),
		Generated: p.Generated(),
		Redirect:  p.Redirect(),
		FmtTitle:  p.FmtTitle(),
		Title:     p.Title(),
		Author:    p.Author(),
	}
	mod, create := p.Modified(), p.Created()
	if !mod.IsZero() {
		info.Modified = &mod
	}
	if !create.IsZero() {
		info.Created = &create
	}
	return info
}

func (p *Page) mainBlock() block {
	return p.main
}

// resets the parser
func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
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
