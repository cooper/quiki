package wiki

import (
	"path/filepath"

	"github.com/cooper/quiki/wikifier"
)

func (w *Wiki) checkDirectories() {
	// TODO
	panic("unimplemented")
}

// relPath returns a path relative to the wiki root
func (w *Wiki) relPath(absPath string) string {
	wikiAbs, _ := filepath.Abs(w.Opt.Dir.Wiki)
	if wikiAbs == "" {
		return ""
	}
	if rel, err := filepath.Rel(wikiAbs, absPath); err == nil {
		return rel
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

// pathForPage returns the absolute path for a page. If necessary, it creates
// diretories for the path components that do not exist.
func (w *Wiki) pathForPage(pageName string, createOK bool, dirPage string) string {
	if dirPage == "" {
		dirPage = w.Opt.Dir.Page
	}
	pageName = wikifier.PageName(pageName)
	if createOK {
		wikifier.MakeDir(dirPage, pageName)
	}
	path, _ := filepath.Abs(filepath.Join(dirPage, pageName))
	return path
}

// pathForCategory returns the absolute path for a category. If necessary, it
// creates directories for the path components that do not exist.
func (w *Wiki) pathForCategory(catName string, catType CategoryType, createOK bool) string {
	catName = wikifier.CategoryName(catName, false)
	dir := filepath.Join(w.Opt.Dir.Cache, "category")
	if createOK {
		wikifier.MakeDir(dir, filepath.Join(string(catType), catName))
	}
	path, _ := filepath.Abs(filepath.Join(dir, string(catType), catName))
	return path
}

// pathForImage returns the absolute path for an image.
func (w *Wiki) pathForImage(imageName string) string {
	path, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Image, imageName))
	return path
}

// pathForModel returns the absolute path for a model.
func (w *Wiki) pathForModel(modelName string) string {
	modelName = wikifier.PageNameExt(modelName, ".model")
	path, _ := filepath.Abs(filepath.Join(w.Opt.Dir.Model, modelName))
	return path
}
