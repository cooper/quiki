package wiki

import (
	"github.com/cooper/quiki/wikifier"
)

func (w *Wiki) allPageFiles() []string {
	files, _ := wikifier.UniqueFilesInDir(w.Opt.Dir.Page, []string{"page"}, false)
	return files
}

func (w *Wiki) allCategoryFiles(catType string) []string {
	dir := w.Opt.Dir.Category
	if catType != "" {
		dir += "/" + catType
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

func (w *Wiki) allMarkdownFiles() []string {
	files, _ := wikifier.UniqueFilesInDir(w.Opt.Dir.Markdown, []string{"md"}, false)
	return files
}
