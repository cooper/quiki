package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newDeleteTestWiki(t *testing.T) *Wiki {
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
	w, err := NewWiki(dir)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func lastCommitMessage(t *testing.T, w *Wiki) string {
	t.Helper()
	repo, err := w.repo()
	if err != nil {
		t.Fatal(err)
	}
	ref, err := repo.Head()
	if err != nil {
		t.Fatal(err)
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(commit.Message)
}

// deleting a page removes the file, its generated cache, and commits
func TestDeletePage(t *testing.T) {
	w := newDeleteTestWiki(t)
	commit := CommitOpts{Name: "tester", Email: "t@example.com"}

	// the first write to a fresh wiki creates the repository, which must
	// not be reported as a failure
	name, err := w.CreatePage("", "Doomed", []byte("@page.title: Doomed;\np { bye }\n"), commit)
	if err != nil {
		t.Fatalf("CreatePage: %v", err)
	}

	// generate the cache
	if _, ok := w.DisplayPage(name).(DisplayPage); !ok {
		t.Fatalf("DisplayPage(%s) failed", name)
	}
	page := w.FindPage(name)
	if _, err := os.Stat(page.CachePath()); err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}

	commit.Comment = "Delete page: " + name
	if err := w.DeletePage(name, commit); err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
	if _, err := os.Stat(w.PathForPage(name)); !os.IsNotExist(err) {
		t.Errorf("page file still exists after delete")
	}
	for _, path := range []string{page.CachePath(), page.SearchPath()} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s still exists after delete", filepath.Base(path))
		}
	}
	if msg := lastCommitMessage(t, w); msg != "Delete page: "+name {
		t.Errorf("last commit %q, want %q", msg, "Delete page: "+name)
	}

	// deleting again fails cleanly
	if err := w.DeletePage(name, commit); err == nil {
		t.Errorf("expected an error deleting a missing page")
	}
}

// deleting an image also removes its generated sizes
func TestDeleteImage(t *testing.T) {
	w := newDeleteTestWiki(t)
	commit := CommitOpts{Name: "tester", Email: "t@example.com"}

	if err := w.WriteImage("pic.png", []byte("not really a png"), true, commit); err != nil {
		t.Fatalf("WriteImage: %v", err)
	}
	cacheDir := filepath.Join(w.Opt.Dir.Cache, "image")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"100x50-pic.png", "200x100-pic@2x.png", "100x50-other.png"} {
		if err := os.WriteFile(filepath.Join(cacheDir, f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := w.DeleteImage("pic.png", commit); err != nil {
		t.Fatalf("DeleteImage: %v", err)
	}
	if _, err := os.Stat(w.PathForImage("pic.png")); !os.IsNotExist(err) {
		t.Errorf("image still exists after delete")
	}
	for _, f := range []string{"100x50-pic.png", "200x100-pic@2x.png"} {
		if _, err := os.Stat(filepath.Join(cacheDir, f)); !os.IsNotExist(err) {
			t.Errorf("generated %s still exists after delete", f)
		}
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "100x50-other.png")); err != nil {
		t.Errorf("unrelated generated image was removed")
	}
}
