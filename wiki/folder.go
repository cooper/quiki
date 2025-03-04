package wiki

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/wikifier"
)

// CreatePageFolder creates a new page folder.
func (w *Wiki) CreatePageFolder(where, name string) (string, error) {
	return w.createFolder(w.Opt.Dir.Page, where, name)
}

// CreateModelFolder creates a new model folder.
func (w *Wiki) CreateModelFolder(where, name string) (string, error) {
	return w.createFolder(w.Opt.Dir.Model, where, name)
}

// CreateImageFolder creates a new image folder.
func (w *Wiki) CreateImageFolder(where, name string) (string, error) {
	return w.createFolder(w.Opt.Dir.Image, where, name)
}

func (w *Wiki) createFolder(base, where, name string) (string, error) {
	name = wikifier.PageNameLink(strings.Replace(name, "/", "_", -1))
	return name, os.MkdirAll(filepath.Join(base, where, name), 0755)
}
