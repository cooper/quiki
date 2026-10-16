package markdown

import (
	"strings"
	"testing"
)

// definition lists become definitions{} with "term: definition;" entries
func TestDefinitionList(t *testing.T) {
	tests := []struct {
		name, in string
		want     []string
	}{
		{"simple", "Term one\n: first definition\n\nTerm two\n: second definition\n",
			[]string{"definitions {", "Term one: first definition;", "Term two: second definition;", "}"}},
		{"multiple definitions", "Term\n: one\n: two\n",
			[]string{"Term: one [br] two;"}},
		{"escapes", "A: colon; semicolon\n: value; with semicolon\n",
			[]string{`A\: colon\; semicolon: value\; with semicolon;`}},
		{"formatting", "Term\n: *emphasis*\n",
			[]string{"Term: [i]emphasis[/i];"}},
	}
	for _, tt := range tests {
		out := string(Run([]byte(tt.in)))
		for _, want := range tt.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s: output missing %q:\n%s", tt.name, want, out)
			}
		}
	}

	// ordinary lists are unchanged
	out := string(Run([]byte("- plain\n- list\n")))
	if !strings.Contains(out, "list {\nplain;\nlist;\n}") {
		t.Errorf("plain list changed:\n%s", out)
	}
}
