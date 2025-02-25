package wiki

import (
	"os"
	"sort"

	"github.com/cooper/quiki/wikifier"
)

// Models returns info about all the models in the wiki.
func (w *Wiki) Models() []wikifier.ModelInfo {
	modelNames := w.allModelFiles()
	models := make([]wikifier.ModelInfo, len(modelNames))

	// models individually
	i := 0
	for _, name := range modelNames {
		models[i] = w.ModelInfo(name)
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

	// convert to []Sortable
	models := w.Models()
	sorted := make([]Sortable, len(models))
	for i, pi := range w.Models() {
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
