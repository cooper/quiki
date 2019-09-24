package wikifier

import (
	"bufio"
	"os"
)

type Page struct {
	FilePath string
	VarsOnly bool
	parser   *parser
	main     block
	*variableScope
}

func NewPage(filePath string) *Page {
	return &Page{FilePath: filePath, variableScope: newVariableScope()}
}

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

// for attributedObject
func (p *Page) MainBlock() block {
	return p.main
}

func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
	p.parser = nil
}

func (p *Page) HTML() HTML {
	// TODO: cache and then recursively destroy elements
	return generateBlock(p.main, p)
}
