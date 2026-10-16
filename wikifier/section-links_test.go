package wikifier

import (
	"strings"
	"testing"
)

func warningsOf(t *testing.T, page *Page) []string {
	t.Helper()
	if err := page.Parse(); err != nil {
		t.Fatal(err)
	}
	page.HTML()
	var msgs []string
	for _, w := range page.Warnings {
		msgs = append(msgs, w.Message)
	}
	return msgs
}

func hasWarning(msgs []string, substr string) bool {
	for _, m := range msgs {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}

// links to sections that exist (before or after the link, or deduplicated)
// are fine; links to sections that don't exist warn
func TestSectionLinkWarnings(t *testing.T) {
	page := NewPageSource(`
Jump to [[ #Later Section ]], [[ #Dup-1 ]] and [[ #Nowhere ]].
sec [Later Section] {
    back to [[ #Later Section ]] and [[ also missing | #Missing Too ]]
}
sec [Dup] { a }
sec [Dup] { b }
`)
	msgs := warningsOf(t, page)

	for _, bad := range []string{"#Nowhere", "#Missing_Too"} {
		if !hasWarning(msgs, "section "+bad+" ") {
			t.Errorf("expected a warning for %s, got %q", bad, msgs)
		}
	}
	for _, good := range []string{"#Later_Section", "#Dup-1"} {
		if hasWarning(msgs, "section "+good+" ") {
			t.Errorf("unexpected warning for existing section %s: %q", good, msgs)
		}
	}
	if n := len(msgs); n != 2 {
		t.Errorf("expected exactly 2 warnings, got %d: %q", n, msgs)
	}
}

// links to other pages' sections can't be checked here and must not warn
func TestSectionLinkOtherPageNotChecked(t *testing.T) {
	page := NewPageSource("see [[ other#Section ]] and [[ /abs/page#Thing ]]\n")
	if msgs := warningsOf(t, page); len(msgs) != 0 {
		t.Errorf("unexpected warnings: %q", msgs)
	}
}

// anchors written in raw html{} or fmt{} content count as existing sections
func TestSectionLinkHTMLAnchors(t *testing.T) {
	page := NewPageSource(`
html { <h2 id="qa-custom">Custom</h2><a name="plain"></a> }
fmt { <span id='inline'>x</span> }
p { [[ #custom ]] [[ #plain ]] [[ #inline ]] [[ #nope ]] }
`)
	msgs := warningsOf(t, page)
	for _, good := range []string{"#custom ", "#plain ", "#inline "} {
		if hasWarning(msgs, "section "+good) {
			t.Errorf("unexpected warning for html anchor %s: %q", good, msgs)
		}
	}
	if !hasWarning(msgs, "section #nope ") {
		t.Errorf("expected a warning for #nope, got %q", msgs)
	}
}

// markdown pages get github-style heading ids and are checked the same way
func TestSectionLinkWarningsMarkdown(t *testing.T) {
	page := NewPageSource("# Top\n\nSee [ok](#second-part) and [bad](#nope).\n\n## Second Part\n\ntext\n")
	page.Markdown = true
	msgs := warningsOf(t, page)
	if !hasWarning(msgs, "section #nope ") {
		t.Errorf("expected a warning for #nope, got %q", msgs)
	}
	if hasWarning(msgs, "section #second-part ") {
		t.Errorf("unexpected warning for existing section: %q", msgs)
	}
}
