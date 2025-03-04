package wiki

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/cooper/quiki/adminifier/utils"
	"github.com/cooper/quiki/wikifier"
)

// Models returns info about all the models in the wiki.
func (w *Wiki) Models() []wikifier.ModelInfo {
	modelNames := w.allModelFiles()
	return w.modelsIn("", modelNames)
}

// ModelsInDir returns info about all the models in the specified directory.
func (w *Wiki) ModelsInDir(where string) []wikifier.ModelInfo {
	modelNames := w.modelFilesInDir(where)
	return w.modelsIn(where, modelNames)
}

func (w *Wiki) modelsIn(prefix string, modelNames []string) []wikifier.ModelInfo {
	models := make([]wikifier.ModelInfo, len(modelNames))
	i := 0
	for _, name := range modelNames {
		models[i] = w.ModelInfo(prefix + name)
		i++
	}
	return models
}

type sortableModelInfo wikifier.ModelInfo

func (mi sortableModelInfo) SortInfo() SortInfo {
	return SortInfo{
		Title: mi.File,
		// TODO: Author
		Created:  *mi.Created,
		Modified: *mi.Modified,
	}
}

// ModelsSorted returns info about all the models in the wiki, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, and SortModified.
func (w *Wiki) ModelsSorted(descend bool, sorters ...SortFunc) []wikifier.ModelInfo {
	return _modelsSorted(w.Models(), descend, sorters...)
}

func _modelsSorted(models []wikifier.ModelInfo, descend bool, sorters ...SortFunc) []wikifier.ModelInfo {

	// convert to []Sortable
	sorted := make([]Sortable, len(models))
	for i, pi := range models {
		sorted[i] = sortableModelInfo(pi)
	}

	// sort
	var sorter sort.Interface = sorter(sorted, sorters...)
	if descend {
		sorter = sort.Reverse(sorter)
	}
	sort.Sort(sorter)

	// convert back to []wikifier.ModelInfo
	for i, si := range sorted {
		models[i] = wikifier.ModelInfo(si.(sortableModelInfo))
	}

	return models
}

// ModelsAndDirs returns info about all the models and directories in a directory.
func (w *Wiki) ModelsAndDirs(where string) ([]wikifier.ModelInfo, []string) {
	models := w.ModelsInDir(where)
	dirs := utils.DirsInDir(filepath.Join(w.Opt.Dir.Model, where))
	return models, dirs
}

// ModelsAndDirsSorted returns info about all the models and directories in a directory, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, and SortModified.
// Directories are always sorted alphabetically (but still respect the descend flag).
func (w *Wiki) ModelsAndDirsSorted(where string, descend bool, sorters ...SortFunc) ([]wikifier.ModelInfo, []string) {
	models, dirs := w.ModelsAndDirs(where)
	models = _modelsSorted(models, descend, sorters...)
	if descend {
		sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	} else {
		sort.Strings(dirs)
	}
	return models, dirs
}

// ModelMap returns a map of model name to wikifier.ModelInfo for all models in the wiki.
func (w *Wiki) ModelMap() map[string]wikifier.ModelInfo {
	modelNames := w.allModelFiles()
	models := make(map[string]wikifier.ModelInfo, len(modelNames))

	// models individually
	for _, name := range modelNames {
		models[name] = w.ModelInfo(name)
	}

	return models
}

// ModelInfo is an inexpensive request for info on a model. It uses cached
// metadata rather than generating the model and extracting variables.
func (w *Wiki) ModelInfo(name string) (info wikifier.ModelInfo) {

	// the model does not exist
	path := w.PathForModel(name)
	mdFi, err := os.Stat(path)
	if err != nil {
		return
	}

	// find model category
	modelCat := w.GetSpecialCategory(name, CategoryTypeModel)

	// if model category exists use that info
	if modelCat.Exists() && modelCat.ModelInfo != nil {
		info = *modelCat.ModelInfo
	}

	// this stuff is available to all
	mod := mdFi.ModTime()
	info.File = name
	info.Path = path
	info.Modified = &mod // actual model mod time

	// fallback title to name
	if info.Title == "" {
		info.Title = wikifier.PageNameNE(name)
	}

	// fallback created to modified
	if info.Created == nil {
		info.Created = &mod
	}

	return
}
