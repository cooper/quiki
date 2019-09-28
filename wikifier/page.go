package wikifier

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// Page represents a single page or article, generally associated with a .page file.
// It provides the most basic public interface to parsing with the wikifier engine.
type Page struct {
	Source   string   // source content
	FilePath string   // Path to the .page file
	VarsOnly bool     // True if Parse() should only extract variables
	Opt      PageOpts // page options
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
	if err := InjectPageOpts(p, &p.Opt); err != nil {
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

// resets the parser
func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
	p.parser = nil
}
