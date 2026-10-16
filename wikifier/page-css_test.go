package wikifier

import (
	"regexp"
	"strings"
	"testing"
)

// every selector generated for a style{} must refer to a class that actually
// appears in the generated HTML, otherwise the style silently does nothing
func TestStyleSelectorsMatchElements(t *testing.T) {
	page := NewPageSource(`
style { background: yellow; }
sec [One] {
    style { color: red; }
    style { font-weight: bold; }
    p { text }
}
sec [Two] {
    style [p] { font-size: 20px; }
    p { text }
}
`)
	page.name = "styled"
	if err := page.Parse(); err != nil {
		t.Fatal(err)
	}
	html := string(page.HTML())
	css := page.CSS()

	// each selector item like .q-main-styled or .q-sec-styled-1 must be an emitted class
	for _, item := range regexp.MustCompile(`\.q-[a-z]+-[\w-]+`).FindAllString(css, -1) {
		class := strings.TrimPrefix(item, ".")
		if !strings.Contains(html, `class="`+class+` `) && !strings.Contains(html, ` `+class+` `) && !strings.Contains(html, ` `+class+`"`) {
			t.Errorf("selector %s refers to class %q which is not in the HTML:\n%s\n%s", item, class, css, html)
		}
	}

	// a style on the main block applies to the main element alone, not scoped twice
	if !strings.Contains(css, ".q-main-styled {") {
		t.Errorf("expected a bare .q-main-styled rule, got:\n%s", css)
	}
	if strings.Contains(css, ".q-main-styled .q-main-styled") {
		t.Errorf("main style is scoped under itself:\n%s", css)
	}

	// two styles on the same section must agree on the section's id
	secRules := regexp.MustCompile(`\.q-main-styled \.q-sec-styled-1 \{`).FindAllString(css, -1)
	if len(secRules) != 2 {
		t.Errorf("expected both section styles to target .q-sec-styled-1, got:\n%s", css)
	}

	// a named style is scoped under its parent
	if !strings.Contains(css, ".q-main-styled .q-sec-styled-2 .q-p {") {
		t.Errorf("expected named style scoped under section, got:\n%s", css)
	}
}

// a nameless page (stdin, NewPageSource) has a main element with no id;
// styles below it are scoped under plain .q-main rather than a dead ".q-main- "
func TestNamelessPageStyles(t *testing.T) {
	page := NewPageSource("sec [One] { style { color: blue; } p { hi } }\n")
	if err := page.Parse(); err != nil {
		t.Fatal(err)
	}
	html := string(page.HTML())
	css := page.CSS()
	if !strings.Contains(html, `class="q-main"`) {
		t.Fatalf("expected an id-less main element:\n%s", html)
	}
	if !strings.Contains(css, ".q-main .q-sec--1 {") {
		t.Errorf("expected .q-main .q-sec--1, got:\n%s", css)
	}

	// calling CSS twice must give the same answer
	if again := page.CSS(); again != css {
		t.Errorf("CSS() is not stable:\n%s\n---\n%s", css, again)
	}
}

// HTMLAndCSS must include the styles, which only exist once the HTML is generated
func TestHTMLAndCSSIncludesStyles(t *testing.T) {
	page := NewPageSource("style { color: red; }\np { hi }\n")
	if err := page.Parse(); err != nil {
		t.Fatal(err)
	}
	out := string(page.HTMLAndCSS())
	if !strings.HasPrefix(out, "<style>") || !strings.Contains(out, "color: red;") {
		t.Errorf("HTMLAndCSS is missing the style block:\n%s", out)
	}
}
