package wikifier

type parserBlock struct {
	parser             *parser
	typ, name          string
	classes            []string
	line, column       int
	endLine, endColumn int
	closed             bool
	parent             *parserBlock
	catch              *parserCatch
}
