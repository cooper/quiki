package wikifier

import "time"

// ModelInfo represents metadata associated with a model.
type ModelInfo struct {
	Title       string     `json:"title"`            // @model.title
	Author      string     `json:"author,omitempty"` // @model.author
	Description string     `json:"desc,omitempty"`   // @model.desc
	File        string     `json:"file"`             // filename
	FileNE      string     `json:"file_ne"`          // filename with no extension
	Path        string     `json:"path"`
	Created     *time.Time `json:"created,omitempty"`  // creation time
	Modified    *time.Time `json:"modified,omitempty"` // modify time
}

// modelInfo is like (wikifier.Page).Info() but used internally
// to instead return a ModelInfo
func (p *Page) modelInfo() ModelInfo {
	info := ModelInfo{
		File:        p.Name(),
		FileNE:      p.NameNE(),
		Title:       p.Title(),
		Author:      p.Author(),
		Description: p.Description(),
	}
	mod, create := p.Modified(), p.Created()
	if !mod.IsZero() {
		info.Modified = &mod
		info.Created = &mod // fallback
	}
	if !create.IsZero() {
		info.Created = &create
	}
	return info
}
