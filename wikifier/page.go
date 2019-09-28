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

// CSS generates and returns the CSS code for the page's inline styles.
func (p *Page) CSS() string {
	generated := ""

	for _, style := range p.styles {
		applyTo := p.cssApplyString(style.applyTo)
		generated += applyTo + " {\n"
		for rule, value := range style.rules {
			generated += "    " + rule + ": " + value + ";\n"
		}
		generated += "}\n"
	}

	return generated
}

func (p *Page) cssApplyString(sets [][]string) string {
	parts := make([]string, len(sets))

	for i, set := range sets {
		str := p.cssSetString(set)
		var start string
		if len(str) > 9 {
			start = str[:9]
		}
		if start == "" || start != ".q-main" {
			id := p.main.el().id()
			str = ".q-" + id + " " + str
		}
		parts[i] = str
	}

	return strings.Join(parts, ",\n")
}

func (p *Page) cssSetString(set []string) string {
	for i, item := range set {
		set[i] = p.cssItemString([]rune(item))
	}
	return strings.Join(set, " ")
}

func (p *Page) cssItemString(chars []rune) string {
	var str string
	var inClass, inID, inElType bool
	for _, char := range chars {

		switch char {

		// starting a class
		case '.':
			inClass = true
			str += ".qc-"

		// starting an ID
		case '#':
			inID = true
			str += ".qi-"

		// in neither a class nor an element type
		// assume that this is the start of element type
		default:
			if !inClass && !inID && !inElType && char != '*' {
				inElType = true
				str += ".q-"
			}
			str += string(char)

		}
	}

	// return $string;
	return str
}

// resets the parser
func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
	p.parser = nil
}
