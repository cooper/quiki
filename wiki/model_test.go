package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestWiki creates a minimal wiki in a temp dir with the given models and pages
func newTestWiki(t *testing.T, models, pages map[string]string) *Wiki {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"pages", "models", "images", "cache", "categories"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "wiki.conf"), []byte("@name: test;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for name, src := range models {
		if err := os.WriteFile(filepath.Join(dir, "models", name), []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for name, src := range pages {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, "pages", name)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "pages", name), []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := NewWiki(dir)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func displayHTMLAndCSS(t *testing.T, w *Wiki, name string) (string, string) {
	t.Helper()
	switch r := w.DisplayPage(name).(type) {
	case DisplayPage:
		return string(r.Content), r.CSS
	default:
		t.Fatalf("DisplayPage(%s) = %#v", name, r)
		return "", ""
	}
}

// styles inside a model must end up in the CSS of the page that uses it,
// and a style on the model root must have a rendered element to apply to
func TestModelStylesReachPage(t *testing.T) {
	w := newTestWiki(t,
		map[string]string{
			"box.model": "style { border: 1px solid red; }\np { style { color: blue; } [@m.text] }\n",
		},
		map[string]string{
			"uses.page": "@page.title: Uses;\n$box { text: hi; }\n",
		},
	)
	html, css := displayHTMLAndCSS(t, w, "uses")

	if !strings.Contains(css, "border: 1px solid red;") || !strings.Contains(css, "color: blue;") {
		t.Fatalf("model styles missing from page CSS:\n%s", css)
	}

	// the model root style targets the model's main element, which must be rendered
	if !strings.Contains(css, ".q-main-model-box {") {
		t.Errorf("expected a .q-main-model-box rule for the model root style:\n%s", css)
	}
	if !strings.Contains(html, "q-main-model-box") {
		t.Errorf("model root element not rendered although a style targets it:\n%s", html)
	}

	// the inner style is scoped under the page like any other style
	if !strings.Contains(css, ".q-main-uses .q-p-model-box-1 {") {
		t.Errorf("expected inner model style scoped under the page:\n%s", css)
	}
	if !strings.Contains(html, "q-p-model-box-1") {
		t.Errorf("inner styled element missing its id class:\n%s", html)
	}
}

// a page named like a model it uses must not share element ids with it
func TestModelIDsDontCollideWithPage(t *testing.T) {
	w := newTestWiki(t,
		map[string]string{"box.model": "style { border: 1px solid red; }\np { style { color: blue; } m }\n"},
		map[string]string{"box.page": "p { style { color: black; } page }\n$box{}\n"},
	)
	html, css := displayHTMLAndCSS(t, w, "box")
	if !strings.Contains(html, `class="q-main-box q-main"`) {
		t.Fatalf("unexpected page main element:\n%s", html)
	}
	if strings.Contains(css, ".q-main-box {") {
		t.Errorf("model root style applies to the page main:\n%s", css)
	}
	if !strings.Contains(css, ".q-main-box .q-p-box-1 {") || !strings.Contains(css, ".q-main-box .q-p-model-box-1 {") {
		t.Errorf("page and model paragraph styles should target different ids:\n%s", css)
	}
}

// the syntax highlighting stylesheet is emitted once however many
// code blocks there are across the page and its models
func TestModelCodeStylesNotDuplicated(t *testing.T) {
	w := newTestWiki(t,
		map[string]string{"codey.model": "code [go] {\nfmt.Println(\"hi\")\n}\n"},
		map[string]string{"uses.page": "$codey{}\n$codey{}\ncode [go] {\nfmt.Println(\"page\")\n}\n"},
	)
	_, css := displayHTMLAndCSS(t, w, "uses")
	if n := strings.Count(css, "/* PreWrapper */"); n != 1 {
		t.Errorf("chroma stylesheet emitted %d times, want 1", n)
	}
}

// pages in subdirectories have "/" in their ids, which must be escaped in selectors
func TestSubdirectoryPageStyles(t *testing.T) {
	w := newTestWiki(t,
		nil,
		map[string]string{"sub/deep.page": "style { color: red; }\nsec [S] { style { color: blue; } p { hi } }\n"},
	)
	html, css := displayHTMLAndCSS(t, w, "sub/deep")
	if !strings.Contains(html, `class="q-main-sub/deep q-main"`) {
		t.Fatalf("unexpected main element:\n%s", html)
	}
	if !strings.Contains(css, `.q-main-sub\/deep {`) || !strings.Contains(css, `.q-main-sub\/deep .q-sec-sub\/deep-1 {`) {
		t.Errorf("expected escaped selectors:\n%s", css)
	}
}

// a model without a root style keeps its default tagless rendering
func TestModelWithoutRootStyleStaysTagless(t *testing.T) {
	w := newTestWiki(t,
		map[string]string{"plain.model": "p { plain }\n"},
		map[string]string{"uses.page": "$plain{}\n"},
	)
	html, css := displayHTMLAndCSS(t, w, "uses")
	if strings.Contains(html, "q-main-plain") {
		t.Errorf("tagless model rendered a wrapper element:\n%s", html)
	}
	if css != "" {
		t.Errorf("unexpected CSS: %s", css)
	}
}
