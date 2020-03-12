package wiki

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cooper/quiki/wikifier"
)

// DisplayFile represents a plain text file to display.
type DisplayFile struct {

	// file name relative to wiki root.
	// path delimiter '/' is always used, regardless of OS.
	File string `json:"file,omitempty"`

	// absolute file path of the file.
	// OS-specific path delimiter is used.
	Path string `json:"path,omitempty"`

	// the plain text file content
	Content string

	// time when the file was last modified
	Modified *time.Time `json:"modified,omitempty"`
}

// DisplayFile returns the display result for a plain text file.
func (w *Wiki) DisplayFile(path string) interface{} {
	var r DisplayFile
	path = filepath.FromSlash(path) // just in case

	// ensure it can be made relative to dir.wiki
	relPath := w.RelPath(path)
	if relPath == "" {
		return DisplayError{
			Error:         "Bad filepath",
			DetailedError: "File '" + path + "' cannot be made relative to '" + w.Opt.Dir.Wiki + "'",
		}
	}

	// file does not exist or can't be read
	fi, err := os.Lstat(path)
	if err != nil {
		return DisplayError{
			Error:         "File does not exist.",
			DetailedError: "File '" + path + "' error: " + err.Error(),
		}
	}

	// read file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return DisplayError{
			Error:         "Error reading file.",
			DetailedError: "File '" + path + "' error: " + err.Error(),
		}
	}

	// results
	mod := fi.ModTime()
	r.File = filepath.ToSlash(relPath)
	r.Path = path
	r.Modified = &mod
	r.Content = string(content)

	return r
}

func (w *Wiki) checkDirectories() {
	// TODO
	panic("unimplemented")
}

// RelPath takes an absolute file path and attempts to make it relative
// to the wiki directory, regardless of whether the path exists.
//
// If the path can be made relative without following symlinks, this is
// preferred. If that fails, symlinks in absPath are followed and a
// second attempt is made.
//
// In any case the path cannot be made relative to the wiki directory,
// an empty string is returned.
func (w *Wiki) RelPath(absPath string) string {
	rel := w._relPath(absPath)
	if strings.Contains(rel, ".."+string(os.PathSeparator)) {
		return ""
	}
	if strings.Contains(rel, string(os.PathSeparator)+"..") {
		return ""
	}
	return rel
}

func (w *Wiki) _relPath(absPath string) string {
	wikiAbs := w.Dir()

	// can't resolve wiki path
	if wikiAbs == "" {
		return ""
	}

	// made it relative as-is
	if rel, err := filepath.Rel(wikiAbs, absPath); err == nil {
		return rel
	}

	// try to make it relative by resolving absPath as absolute
	absPath, _ = filepath.Abs(absPath)
	if absPath != "" {
		if rel, err := filepath.Rel(wikiAbs, absPath); err == nil {
			return rel
		}
	}

	return ""
}

func (w *Wiki) allPageFiles() []string {
	files, _ := wikifier.UniqueFilesInDir(w.Opt.Dir.Page, []string{"page", "md"}, false)
	return files
}

func (w *Wiki) allCategoryFiles(catType CategoryType) []string {
	dir := w.Opt.Dir.Category
	if catType != "" {
		dir = filepath.Join(dir, string(catType))
	}
	files, _ := wikifier.UniqueFilesInDir(dir, []string{"cat"}, false)
	return files
}

func (w *Wiki) allModelFiles() []string {
	files, _ := wikifier.UniqueFilesInDir(w.Opt.Dir.Model, []string{"model"}, false)
	return files
}

func (w *Wiki) allImageFiles() []string {
	files, _ := wikifier.UniqueFilesInDir(w.Opt.Dir.Image, []string{"png", "jpg", "jpeg"}, false)
	return files
}

// pathForPage returns the absolute path for a page.
func (w *Wiki) pathForPage(pageName string) string {

	// try lowercased version first (quiki style)
	lcPageName := filepath.FromSlash(wikifier.PageName(pageName))
	path, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Page, lcPageName))

	// it doesn't exist; try non-lowercased version (markdown/etc)
	if _, err := os.Stat(path); err != nil {
		normalPageName := filepath.FromSlash(wikifier.PageName(pageName))
		normalPath, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Page, normalPageName))
		if _, err := os.Stat(normalPath); err == nil {
			return normalPath
		}
	}

	return path
}

// pathForCategory returns the absolute path for a category. If necessary, it
// creates directories for the path components that do not exist.
func (w *Wiki) pathForCategory(catName string, catType CategoryType, createOK bool) string {
	catName = wikifier.CategoryName(catName)
	dir := filepath.Join(w.Opt.Dir.Cache, "category")
	if createOK {
		wikifier.MakeDir(dir, filepath.Join(string(catType), catName))
	}
	path, _ := filepath.Abs(filepath.Join(dir, string(catType), catName))
	return path
}

// pathForImage returns the absolute path for an image.
func (w *Wiki) pathForImage(imageName string) string {
	path, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Image, filepath.FromSlash(imageName)))
	return path
}

// pathForModel returns the absolute path for a model.
func (w *Wiki) pathForModel(modelName string) string {
	modelName = wikifier.PageNameExt(modelName, ".model")
	path, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Model, filepath.FromSlash(modelName)))
	return path
}

// Dir returns the absolute path to the resolved wiki directory.
// If the wiki directory is a symlink, it is followed.
//
// Optional path components can be passed as arguments to be joined
// with the wiki root by the path separator.
func (w *Wiki) Dir(dirs ...string) string {
	wikiAbs, _ := filepath.Abs(w.Opt.Dir.Wiki)
	return filepath.Join(append([]string{wikiAbs}, dirs...)...)
}

// UnresolvedAbsFilePath takes a relative path to a file within the wiki
// (e.g. `pages/mypage.page`) and joins it with the absolute path to the wiki
// directory. The result is an absolute path which may or may not exist.
//
// Symlinks are not followed. If that is desired, use absoluteFilePath instead.
//
func (w *Wiki) UnresolvedAbsFilePath(relPath string) string {

	// sanitize
	relPath = filepath.FromSlash(relPath)

	// join with wiki dir
	path := w.Dir(relPath)

	// resolve symlink
	abs, _ := filepath.Abs(path)
	return abs
}

// AbsFilePath takes a relative path to a file within the wiki
// (e.g. `pages/mypage.page`), joins it with the wiki directory, and evaluates it
// with `filepath.Abs()`. The result is an absolute path which may or may not exist.
//
// If the file is a symlink, it is followed. Thus, it is possible for the resulting
// path to exist outside the wiki directory. If that is not desired, use
// unresolvedAbsFilePath instead.
func (w *Wiki) AbsFilePath(relPath string) string {
	path, _ := filepath.Abs(w.Dir(w.UnresolvedAbsFilePath(relPath)))
	return path
}
