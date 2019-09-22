package wikifier

import (
	"bufio"
	"os"
)

type Page struct {
	filePath string
	parser   *parser
}

func NewPage(filePath string) *Page {
	return &Page{filePath, newParser()}
}

func (p *Page) Parse() error {
	file, err := os.Open(p.filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p.parser.parseLine(scanner.Bytes())
	}

	return scanner.Err()
}

func (p *Page) HTML() Html {
	return generateBlock(p.parser.block)
}
