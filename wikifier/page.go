package wikifier

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Page represents a single page or article, generally associated with a .page file.
// It provides the most basic public interface to parsing with the wikifier engine.
type Page struct {
	Source   string  // source content
	FilePath string  // Path to the .page file
	VarsOnly bool    // True if Parse() should only extract variables
	Opt      PageOpt // page options
	styles   []styleEntry
	parser   *parser // wikifier parser instance
	main     block   // main block
	images   map[string][][]int
	*variableScope
}

// NewPage creates a page given its filepath.
func NewPage(filePath string) *Page {
	return &Page{FilePath: filePath, Opt: defaultPageOpt, variableScope: newVariableScope()}
}

// NewPageSource creates a page given some source code.
func NewPageSource(source string) *Page {
	return &Page{Source: source, Opt: defaultPageOpt, variableScope: newVariableScope()}
}

// Parse opens the page file and attempts to parse it, returning any errors encountered.
func (p *Page) Parse() error {
	p.parser = newParser()
	p.main = p.parser.block

	// create reader from file path or source code provided
	var reader io.Reader
	if p.Source != "" {
		reader = strings.NewReader(p.Source)
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

	// parse the blocks, unless we only want vars
	if !p.VarsOnly {
		p.main.parse(p)
	}

	// inject variables set in the page to page opts
	if err := InjectPageOpt(p, &p.Opt); err != nil {
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

// # page filename, with or without extension.
// # this DOES take symbolic links into account.
// # this DOES include the page prefix, if applicable.
// sub name {
//     my $page = shift;
//     return $page->{abs_name}
//         if length $page->{abs_name};
//     return $page->{cached_props}{name} //= do {
//         my $dir  = $page->opt('dir.page');
//         my $path = $page->path;
//         (my $name = $path) =~ s/^\Q$dir\E(\/?)//;
//         index($path, $dir) ? $page->rel_name : $name;
//     };
// }

// Name returns the resolved page name, with or without extension.
//
// This does NOT take symbolic links into account.
// It DOES include the page prefix, however, if applicable.
//
func (p *Page) Name() string {
	dir := p.Opt.Dir.Page
	path := p.Path()
	name := strings.TrimPrefix(path, dir)
	name = strings.TrimPrefix(name, "/")
	if strings.Index(path, dir) != -1 {
		return filepath.Base(p.RelPath())
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
	dir := strings.TrimSuffix(filepath.Dir(p.Name()), "/")
	if dir == "." {
		return ""
	}
	return dir
}

// Path returns the absolute path to the page as resolved.
// If the path does not resolve, returns an empty string.
func (p *Page) Path() string {
	path, _ := filepath.Abs(p.RelPath())
	return path
}

// RelName returns the unresolved page filename, with or without extension.
// This does NOT take symbolic links into account.
// It is not guaranteed to exist.
func (p *Page) RelName() string {
	dir := p.Opt.Dir.Page
	path := p.RelPath()
	name := strings.TrimPrefix(path, dir)
	name = strings.TrimPrefix(name, "/")
	if strings.Index(path, dir) != -1 {
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
	return p.FilePath
}

// resets the parser
func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
	p.parser = nil
}
