package wiki

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/cooper/quiki/wikifier"
)

// CategoryType describes the type of a Category.
type CategoryType string

const (
	// CategoryTypeImage is a type of category that tracks which pages use an image.
	CategoryTypeImage CategoryType = "image"

	// CategoryTypeModel is a type of category that tracks which pages use a model.
	CategoryTypeModel = "model"

	// CategoryTypePage is a type of category that tracks which pages reference another page.
	CategoryTypePage = "page"
)

// A Category is a collection of pages pertaining to a topic.
//
// A page can belong to many categories. Category memberships and metadta
// are stored in JSON manifests.
//
type Category struct {

	// category path
	Path string `json:"-"`

	// category filename, including the .cat extension
	File string `json:"file,omitempty"`

	// category name without extension
	Name string `json:"name,omitempty"`

	// human-readable category title
	Title string `json:"title,omitempty"`

	// time when the category was created
	Created time.Time `json:"created,omitempty"`

	// time when the category was last modified.
	// this is updated when pages are added and deleted
	Modified time.Time `json:"modified,omitempty"`

	// pages in the category. keys are filenames
	Pages map[string]CategoryEntry `json:"pages,omitempty"`

	// when true, the category is preserved even when no pages remain
	Preserve bool `json:"preserve,omitempty"`

	// EXTRAS

	// if applicable, this is the type of special category.
	// for normal categories, this is empty
	Type CategoryType `json:"type,omitempty"`

	// for CategoryTypePage, this is the info for the tracked page
	PageInfo wikifier.PageInfo `json:"page_info,omitempty"`
}

// A CategoryEntry describes a page that belongs to a category.
type CategoryEntry struct {

	// time at which the page metadata in this category file was last updated.
	// this is compared against page file modification time
	Asof time.Time `json:"asof,omitempty"`

	// embedded page info
	// note this info is accurate only as of the Asof time
	wikifier.PageInfo

	// EXTRAS

	// for CategoryTypeImage, an array of image dimensions used on this page.
	// dimensions are guaranteed to be positive integers. the number of elements will
	// always be even, since each occurence of the image produces two (width and then height)
	Dimensions [][]int `json:"dimensions,omitempty"`

	// for CategoryTypePage, an array of line numbers on which the tracked page is
	// referenced on the page described by this entry
	Lines []int `json:"lines,omitempty"`
}

// GetCategory loads or creates a category.
func (w *Wiki) GetCategory(name string) *Category {
	return w.GetSpecialCategory(name, "")
}

// GetSpecialCategory loads or creates a special category given the type.
func (w *Wiki) GetSpecialCategory(name string, typ CategoryType) *Category {
	name = wikifier.CategoryName(name, false)
	path := w.pathForCategory(name, typ, true)

	// load the category if it exists
	var cat Category
	jsonData, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(jsonData, &cat)
	}

	// if an error occurred in reading or parsing, ditch the file
	// note it may or may not exist anyway
	if err != nil {
		log.Printf("GetCategory(%s): %v", name, err)
		os.Remove(path)
	}

	// update these
	cat.Path = path
	cat.Name = name
	cat.Type = typ

	return &cat
}

// AddPage adds a page to a category.
//
// If the page already belongs and any information has changed, the category is updated.
// If force is true,
func (cat *Category) AddPage(page *wikifier.Page) {
	cat.addPageExtras(page, nil, nil)
}

func (cat *Category) addPageExtras(pageMaybe *wikifier.Page, dimensions [][]int, lines []int) {
	if pageMaybe != nil {
		mod := pageMaybe.Modified()
		// TODO: if the page was just renamed, delete the old entry

		// the page has not changed since the asof time, so do nothing
		entry, exist := cat.Pages[pageMaybe.Name()]
		if exist && mod.Before(entry.Asof) || mod.Equal(entry.Asof) {
			return
		}
	}

	// if this is a new category with zero pages, it must have the
	// preserve flag
	if len(cat.Pages) == 0 && pageMaybe == nil && !cat.Preserve {
		panic("attempting to create category with no pages: " + cat.Name)
	}

	// ok, at this point we're gonna add or update the page if there is one
	if pageMaybe != nil {
		if cat.Pages == nil {
			cat.Pages = make(map[string]CategoryEntry)
		}
		cat.Pages[pageMaybe.Name()] = CategoryEntry{
			Asof:       time.Now(),
			PageInfo:   pageMaybe.Info(),
			Dimensions: dimensions,
			Lines:      lines,
		}
	}

	// write it
	cat.write()
}

// write category to file
func (cat *Category) write() {

	// encode as JSON
	jsonData, err := json.Marshal(cat)
	if err != nil {
		log.Printf("Category(%s).write(): %v", cat.Name, err)
		return
	}

	// write
	ioutil.WriteFile(cat.Path, jsonData, 0666)
}

// cat_check_page
func (w *Wiki) updatePageCategories(page *wikifier.Page) {

	// page metadata category
	pageCat := w.GetSpecialCategory(page.Name(), CategoryTypePage)
	pageCat.PageInfo = page.Info()
	pageCat.Preserve = true // keep until page no longer exists
	pageCat.addPageExtras(nil, nil, nil)

	// actual categories
	for _, name := range page.Categories() {
		w.GetCategory(name).AddPage(page)
	}

	// TODO: page, image, and model categories
}
