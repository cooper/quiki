package markdown_test

import (
	"html"
	"regexp"
	"strings"
	"testing"

	"github.com/cooper/quiki/markdown"
	"github.com/cooper/quiki/wikifier"
)

var tagRe = regexp.MustCompile(`<[^>]*>`)

// renderText converts real markdown input to quiki source and then to HTML,
// returning the visible text of the page (tags stripped, entities decoded) so
// that chroma's highlighting spans do not get in the way of comparisons.
func renderText(t *testing.T, md string) string {
	t.Helper()
	p := wikifier.NewPage("test.md")
	p.Source = md
	if err := p.Parse(); err != nil {
		t.Fatalf("parse error for %q: %v", md, err)
	}
	return html.UnescapeString(tagRe.ReplaceAllString(string(p.HTML()), ""))
}

// quiki#22: backslashes in markdown must come out of the rendered page
// exactly as CommonMark would display them, with no extra backslash added
// by the markdown -> quiki source -> HTML round trip.
func TestBackslashes(t *testing.T) {
	tests := []struct {
		name, md string
		want     string // must appear in rendered text
		wantNot  string // must not appear in rendered text
	}{
		{"paragraph", `a\b C:\Users\foo`, `a\b C:\Users\foo`, `\\`},
		{"paragraph escaped backslash", `a\\b`, `a\b`, `\\`},
		{"inline code", "`a\\b`", `a\b`, `\\`},
		{"inline code double backslash", "`a\\\\b`", `a\\b`, `\\\`},
		{"fenced code", "```\na\\b\nc\\\\d\n```\n", "a\\b\nc\\\\d\n", `\\\`},
		{"fenced code with lang", "```sh\necho a\\b\n```\n", "echo a\\b\n", `\\`},
		{"fenced code with braces", "```c\nif (x) {\n  y\\n;\n}\n```\n", "if (x) {\n  y\\n;\n}\n", `\\`},
		{"indented code", "para\n\n    a\\b\n    c\\\\d\n", "a\\b\nc\\\\d\n", `\\\`},
		{"html block", "<div>a\\b</div>\n", `a\b`, `\\`},
		{"heading", "# a\\b\n", `a\b`, `\\`},
		{"list item", "- a\\b\n", `a\b`, `\\`},
		{"link text", `[a\b](http://example.com/)`, `a\b`, `\\`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderText(t, tt.md)
			if !strings.Contains(got, tt.want) {
				t.Errorf("markdown %q\nrendered text %q\nwant it to contain %q", tt.md, got, tt.want)
			}
			if tt.wantNot != "" && strings.Contains(got, tt.wantNot) {
				t.Errorf("markdown %q\nrendered text %q\nmust not contain %q", tt.md, got, tt.wantNot)
			}
		})
	}
}

// Code and HTML blocks whose braces do not balance cannot use brace-escape;
// they must still parse and keep their braces.
func TestUnbalancedBraces(t *testing.T) {
	got := renderText(t, "```\n}\n{\n```\n")
	if !strings.Contains(got, "}\n{\n") {
		t.Errorf("rendered text %q, want it to contain %q", got, "}\n{\n")
	}
	got = renderText(t, "<div>}</div>\n")
	if !strings.Contains(got, "}") {
		t.Errorf("rendered text %q, want it to contain %q", got, "}")
	}
}

// Run should emit brace-escaped blocks when it can, and fall back to the
// escaped single-brace form otherwise.
func TestRunBraceEscape(t *testing.T) {
	tests := []struct{ name, md, want string }{
		{"balanced code", "```js\nf({a: 1});\n```\n", "~code [js] {{\nf({a: 1});\n}}"},
		{"backslashes in code", "```\na\\b\n```\n", "~code {{\na\\b\n}}"},
		{"unbalanced code", "```\n}\n```\n", "~code {\n\\}\n}"},
		{"balanced html", "<div>a\\b</div>\n", "~html {{\n<div>a\\b</div>"},
		{"unbalanced html", "<div>{</div>\n", "~html {\n<div>\\{</div>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(markdown.Run([]byte(tt.md)))
			if !strings.Contains(got, tt.want) {
				t.Errorf("Run(%q) =\n%s\nwant it to contain %q", tt.md, got, tt.want)
			}
		})
	}
}
