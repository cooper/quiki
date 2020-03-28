package wiki

import (
	"os"
	"sort"
	"time"
)

// ModelInfo represents metadata associated with a model.
type ModelInfo struct {
	File     string     `json:"file"` // filename
	Path     string     `json:"path"`
	Created  *time.Time `json:"created,omitempty"`  // creation time
	Modified *time.Time `json:"modified,omitempty"` // modify time
}

// Models returns info about all the models in the wiki.
func (w *Wiki) Models() []ModelInfo {
	modelNames := w.allModelFiles()
	models := make([]ModelInfo, len(modelNames))

	// models individually
	i := 0
	for _, name := range modelNames {
		models[i] = w.ModelInfo(name)
		i++
	}

	return models
}

type sortableModelInfo ModelInfo

func (mi sortableModelInfo) SortInfo() SortInfo {
	return SortInfo{
		Title: mi.File,
		// TODO: Author
		Created:  *mi.Created,
		Modified: *mi.Modified,
	}
}

// ModelsSorted returns info about all the models in the wiki, sorted as specified.
func (w *Wiki) ModelsSorted(descend bool, sorters ...SortFunc) []ModelInfo {

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

	// convert back to []ModelInfo
	for i, si := range sorted {
		models[i] = ModelInfo(si.(sortableModelInfo))
	}

	return models
}

// ModelMap returns a map of model name to ModelInfo for all models in the wiki.
func (w *Wiki) ModelMap() map[string]ModelInfo {
	modelNames := w.allModelFiles()
	models := make(map[string]ModelInfo, len(modelNames))

	// models individually
	for _, name := range modelNames {
		models[name] = w.ModelInfo(name)
	}

	return models
}

// ModelInfo is an inexpensive request for info on a model. It uses cached
// metadata rather than generating the model and extracting variables.
func (w *Wiki) ModelInfo(name string) (info ModelInfo) {

	// the model does not exist
	path := w.pathForModel(name)
	mdFi, err := os.Stat(path)
	if err != nil {
		return
	}

	// find model category
	modelCat := w.GetSpecialCategory(name, CategoryTypeModel)

	// if model category exists use that info
	if modelCat.Exists() {
		// TODO: store/retrieve some metadata in the model category
	}

	// this stuff is available to all
	mod := mdFi.ModTime()
	info.File = name
	info.Path = path
	info.Modified = &mod // actual model mod time
	info.Created = &mod

	return
}
