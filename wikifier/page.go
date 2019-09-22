package wikifier

import (
	"bufio"
	"os"
)

type Page struct {
	FilePath  string
	parser    *parser
	mainBlock block
	*variableScope
}

func NewPage(filePath string) *Page {
	return &Page{FilePath: filePath, variableScope: newVariableScope()}
}

func (p *Page) Parse() error {
	p.parser = newParser()
	p.mainBlock = p.parser.block
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

	return scanner.Err()
}

func (p *Page) resetParseState() {
	// TODO: recursively destroy blocks
	p.parser = nil
}

func (p *Page) HTML() Html {
	// TODO: cache and then recursively destroy elements
	return generateBlock(p.mainBlock, p)
}
