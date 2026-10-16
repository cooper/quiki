package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a model's sections count for the page it is used on, and the model's own
// section links are checked against the whole page
func TestSectionLinkWarningsAcrossModel(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"pages", "models", "images", "cache", "categories"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"wiki.conf":        "@name: test;\n",
		"models/box.model": "\n\nsec [In Model] { back to [[ #On Page ]] and [[ #Also Gone ]] }\n",
		"pages/uses.page":  "sec [On Page] { see [[ #In Model ]] and [[ #Gone ]] }\n$box{}\n",
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
	var msgs []string
	for _, warning := range r.Warnings {
		msgs = append(msgs, warning.Message)
	}
	joined := strings.Join(msgs, "\n")
	for _, good := range []string{"#In_Model ", "#On_Page "} {
		if strings.Contains(joined, "section "+good) {
			t.Errorf("section %s exists across page and model, got %q", good, msgs)
		}
	}
	for _, bad := range []string{"#Gone ", "#Also_Gone "} {
		if !strings.Contains(joined, "section "+bad) {
			t.Errorf("expected a warning for %s, got %q", bad, msgs)
		}
	}

	// the model's warning is reported at the $box{} block in the page (line 2)
	// and says where in the model the link is (line 3)
	for _, warning := range r.Warnings {
		if strings.Contains(warning.Message, "#Also_Gone ") {
			if warning.Pos.Line != 2 {
				t.Errorf("model link warning at page line %d, want 2 (the model block)", warning.Pos.Line)
			}
			if !strings.Contains(warning.Message, "(in model $box at 3:") {
				t.Errorf("model link warning doesn't say where in the model: %q", warning.Message)
			}
		}
	}
}
