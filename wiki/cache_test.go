package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

// purging removes generated pages and images but keeps metadata
func TestPurgeCache(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"pages", "models", "images", "cache", "categories"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "wiki.conf"), []byte("@name: test;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	w, err := NewWiki(dir)
	if err != nil {
		t.Fatal(err)
	}

	cache := w.Opt.Dir.Cache
	for _, rel := range []string{"page/main.page.cache", "page/sub/deep.page.txt", "image/100x100-pic.png", "meta/page/main.page.cat", "category/topic.cat"} {
		path := filepath.Join(cache, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := w.PurgeCache(); err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"page", "image"} {
		entries, err := os.ReadDir(filepath.Join(cache, sub))
		if err != nil {
			t.Errorf("%s cache dir missing after purge: %v", sub, err)
		} else if len(entries) != 0 {
			t.Errorf("%s cache not empty after purge: %v", sub, entries)
		}
	}
	for _, rel := range []string{"meta/page/main.page.cat", "category/topic.cat"} {
		if _, err := os.Stat(filepath.Join(cache, rel)); err != nil {
			t.Errorf("metadata %s removed by purge", rel)
		}
	}
}
