package wikifier

import (
	"bufio"
	"os"
)

// Page represents a single page or article, generally associated with a .page file.
// It provides the most basic public interface to parsing with the wikifier engine.
type Page struct {
	FilePath string
	VarsOnly bool
	parser   *parser
	main     block
	*variableScope
}

// NewPage creates a page given its filepath.
func NewPage(filePath string) *Page {
	return &Page{FilePath: filePath, variableScope: newVariableScope()}
}

// Parse opens the page file and attempts to parse it, returning any errors encountered.
func (p *Page) Parse() error {
	p.parser = newParser()
	p.main = p.parser.block
	file, err := os.Open(p.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := p.parser.parseLine(scanner.Bytes(), p); err != nil {
			return err
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}

	// TODO: check if p.parser.catch != main block

	//  parse the blocks, unless we only want vars
	if !p.VarsOnly {
		p.main.parse(p)
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
