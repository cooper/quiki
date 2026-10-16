package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a model may be written in markdown as name.md and used as $name{}
func TestMarkdownModel(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"pages", "models", "images", "cache", "categories"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"wiki.conf":          "@name: test;\n",
		"models/note.md":     "# Note Title\n\nThis is **markdown** in a model.\n\n## Second\n\nmore\n",
		"models/plain.model": "p { quiki model }\n",
		"pages/uses.page":    "@page.title: Uses;\n$note{}\n$plain{}\n$missing{}\n",
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := NewWiki(dir)
	if err != nil {
		t.Fatal(err)
	}

	r, ok := w.DisplayPage("uses").(DisplayPage)
	if !ok {
		t.Fatalf("DisplayPage: %#v", w.DisplayPage("uses"))
	}
	html := string(r.Content)
	if !strings.Contains(html, `<span style="font-weight: bold;">markdown</span>`) {
		t.Errorf("markdown model not rendered as markdown:\n%s", html)
	}
	if !strings.Contains(html, "quiki model") {
		t.Errorf("quiki model not rendered:\n%s", html)
	}

	// a model is not a page: no injected table of contents or page variables
	if strings.Contains(html, "q-toc") {
		t.Errorf("markdown model rendered a table of contents:\n%s", html)
	}
	if strings.Contains(html, "Markdown") {
		t.Errorf("markdown model leaked page variables into output:\n%s", html)
	}

	// the first heading is the model's title
	if info := w.ModelInfo("note.md"); info.Title != "Note Title" {
		t.Errorf("ModelInfo(note.md).Title = %q, want %q", info.Title, "Note Title")
	}

	// the page records both models by their real file names
	var warned bool
	for _, warning := range r.Warnings {
		warned = warned || strings.Contains(warning.Message, "$missing{} does not exist")
	}
	if !warned {
		t.Errorf("expected a warning for the missing model, got %v", r.Warnings)
	}

	// model listings include markdown models
	var names []string
	for _, info := range w.Models() {
		names = append(names, info.File)
	}
	if !strings.Contains(strings.Join(names, ","), "note.md") || !strings.Contains(strings.Join(names, ","), "plain.model") {
		t.Errorf("model listing missing files: %v", names)
	}
	if info := w.ModelInfo("note.md"); info.File == "" {
		t.Errorf("ModelInfo(note.md) found nothing")
	}
}
