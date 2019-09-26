package wikifier

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Page represents a single page or article, generally associated with a .page file.
// It provides the most basic public interface to parsing with the wikifier engine.
type Page struct {
	Source     string   // source content
	FilePath   string   // Path to the .page file
	VarsOnly   bool     // True if Parse() should only extract variables
	Opt        PageOpts // page options
	styles     []styleEntry
	parser     *parser // wikifier parser instance
	main       block   // main block
	images     map[string][][]int
	imagesFull map[string][][]int
	*variableScope
}

// NewPage creates a page given its filepath.
func NewPage(filePath string) *Page {
	return &Page{FilePath: filePath, variableScope: newVariableScope()}
}

// NewPageSource creates a page given some source code.
func NewPageSource(source string) *Page {
	return &Page{Source: source, variableScope: newVariableScope()}
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
	// my $page = shift;
	// return unless $page->{styles};
	// my $string = '';
	generated := ""

	// foreach my $rule_set (@{ $page->{styles} }) {
	for _, style := range p.styles {

		//     my $apply_to = $page->_css_apply_string(@{ $rule_set->{apply_to} });
		applyTo := p.cssApplyString(style.applyTo)

		//     $string     .= "$apply_to {\n";
		generated += applyTo + " {\n"

		//     foreach my $rule (keys %{ $rule_set->{rules} }) {
		for rule, value := range style.rules {
			generated += "    " + rule + ": " + value + ";\n"
			//         my $value = $rule_set->{rules}{$rule};
			//         $string  .= "    $rule: $value;\n";
		}
		//     $string .= "}\n";
		generated += "}\n"
	}
	// return $string;
	return generated
}

func (p *Page) cssApplyString(sets [][]string) string {
	fmt.Printf("sets: %v\n", sets)
	// my ($page, @sets) = @_;
	// # @sets = an array of [
	// #   ['section'],
	// #   ['.someClass'],
	// #   ['section', '.someClass'],
	// #   ['section', '.someClass.someOther']
	// # ] etc.

	parts := make([]string, len(sets))

	// return join ",\n", map {
	//     my $string = $page->_css_set_string(@$_);
	//     my $start  = substr $string, 0, 10;
	//     if (!$start || $start ne '.wiki-main') {
	//         my $id  = $page->{wikifier}{main_block}{element}{id};
	//         $string = ".wiki-$id $string";
	//     }
	//     $string
	// } @sets;
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
	//    return join ' ', map { $page->_css_item_string(split //, $_) } @items;
	for i, item := range set {
		set[i] = p.cssItemString([]rune(item))
	}
	return strings.Join(set, " ")
}

func (p *Page) cssItemString(chars []rune) string {
	// my ($string, $in_class, $in_id, $in_el_type) = '';
	var str string
	var inClass, inID, inElType bool

	// foreach my $char (@chars) {
	for _, char := range chars {

		switch char {

		//     # we're starting a class.
		//     if ($char eq '.') {
		//         $in_class++;
		//         $string .= '.wiki-class-';
		//         next;
		//     }
		case '.':
			inClass = true
			str += ".qc-"

		//     # we're starting an ID.
		//     if ($char eq '#') {
		//         $in_id++;
		//         $string .= '.wiki-id-';
		//         next;
		//     }
		case '#':
			inID = true
			str += ".qi-"

		default:
			//     # we're in neither a class nor an element type.
			//     # assume that this is the start of element type.
			//     if (!$in_class && !$in_id && !$in_el_type && $char ne '*') {
			//         $in_el_type = 1;
			//         $string .= '.wiki-';
			//     }
			if !inClass && !inID && !inElType && char != '*' {
				inElType = true
				str += ".q-"
			}

			//     $string .= $char;
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
