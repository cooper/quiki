package wiki

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/Songmu/go-httpdate"
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
// A page can belong to many categories. Category memberships and metadata
// are stored in JSON manifests.
//
type Category struct {

	// category path
	Path string `json:"-"`

	// category filename, including the .cat extension
	File string `json:"-"`

	// category name without extension
	Name string `json:"name,omitempty"`

	// human-readable category title
	Title string `json:"title,omitempty"`

	// time when the category was created
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

	// time when the category was last modified.
	// this is updated when pages are added and deleted
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted

	// pages in the category. keys are filenames
	Pages map[string]CategoryEntry `json:"pages,omitempty"`

	// when true, the category is preserved even when no pages remain
	Preserve bool `json:"preserve,omitempty"`

	// EXTRAS

	// if applicable, this is the type of special category.
	// for normal categories, this is empty
	Type CategoryType `json:"type,omitempty"`

	// for CategoryTypePage, this is the info for the tracked page
	PageInfo *wikifier.PageInfo `json:"page_info,omitempty"`

	// for CategoryTypeImage, this is the info for the tracked image
	ImageInfo *struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"image_info,omitempty"`
}

// A CategoryEntry describes a page that belongs to a category.
type CategoryEntry struct {

	// time at which the page metadata in this category file was last updated.
	// this is compared against page file modification time
	Asof *time.Time `json:"asof,omitempty"`

	// embedded page info
	// note this info is accurate only as of the Asof time
	wikifier.PageInfo

	// EXTRAS

	// for CategoryTypeImage, an array of image dimensions used on this page.
	// dimensions are guaranteed to be positive integers. the number of elements will
	// always be even, since each occurrence of the image produces two (width and then height)
	Dimensions [][]int `json:"dimensions,omitempty"`

	// for CategoryTypePage, an array of line numbers on which the tracked page is
	// referenced on the page described by this entry
	Lines []int `json:"lines,omitempty"`
}

// DisplayCategoryPosts represents a category result to display.
type DisplayCategoryPosts struct {

	// DisplayPage results
	// overrides the Category Pages field
	Pages []DisplayPage `json:"pages,omitempty"`

	// the page number (first page = 0)
	PageN int `json:"page_n"`

	// the total number of pages
	NumPages int `json:"num_pages"`

	// this is the combined CSS for all pages we're displaying
	CSS string `json:"css,omitempty"`

	// all other fields are inherited from the category itself
	*Category
}

// GetCategory loads or creates a category.
func (w *Wiki) GetCategory(name string) *Category {
	return w.GetSpecialCategory(name, "")
}

// GetSpecialCategory loads or creates a special category given the type.
func (w *Wiki) GetSpecialCategory(name string, typ CategoryType) *Category {
	name = wikifier.CategoryNameNE(name)
	path := w.pathForCategory(name, typ, true)

	// load the category if it exists
	var cat Category
	jsonData, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(jsonData, &cat)
	} else {
		now := time.Now()
		cat.Created = &now
		cat.Modified = &now
		cat.CreatedHTTP = httpdate.Time2Str(now)
		cat.ModifiedHTTP = cat.CreatedHTTP
		err = nil
	}

	// if an error occurred in parsing, ditch the file
	// note it may or may not exist anyway
	if err != nil {
		w.Debugf("GetCategory(%s): %v", name, err)
		os.Remove(path)
	}

	// update these
	cat.Path = path
	cat.Name = name
	cat.File = name + ".cat"
	cat.Type = typ

	return &cat
}

// AddPage adds a page to a category.
//
// If the page already belongs and any information has changed, the category is updated.
// If force is true,
func (cat *Category) AddPage(w *Wiki, page *wikifier.Page) {
	cat.addPageExtras(w, page, nil, nil)
}

func (cat *Category) addPageExtras(w *Wiki, pageMaybe *wikifier.Page, dimensions [][]int, lines []int) {

	// update existing info
	cat.update(w)

	// do nothing if the entry exists and the page has not changed since the asof time
	if pageMaybe != nil {
		mod := pageMaybe.Modified()
		// TODO: if the page was just renamed, delete the old entry

		// the page has not changed since the asof time, so do nothing
		entry, exist := cat.Pages[pageMaybe.Name()]
		if exist && entry.Asof != nil {
			if mod.Before(*entry.Asof) || mod.Equal(*entry.Asof) {
				return
			}
		}
	}

	// if this is a new category with zero pages, it must have the
	// preserve flag
	if len(cat.Pages) == 0 && pageMaybe == nil && !cat.Preserve {
		panic("attempting to create category with no pages: " + cat.Name)
	}

	// ok, at this point we're gonna add or update the page if there is one
	now := time.Now()
	cat.Modified = &now
	cat.ModifiedHTTP = httpdate.Time2Str(now)
	if pageMaybe != nil {
		if cat.Pages == nil {
			cat.Pages = make(map[string]CategoryEntry)
		}
		cat.Pages[pageMaybe.Name()] = CategoryEntry{
			Asof:       &now,
			PageInfo:   pageMaybe.Info(),
			Dimensions: dimensions,
			Lines:      lines,
		}
	}

	// write it
	cat.write(w)
}

// Exists returns whether a category currently exists.
func (cat *Category) Exists() bool {
	_, err := os.Lstat(cat.Path)
	return err == nil
}

// write category to file
func (cat *Category) write(w *Wiki) {

	// encode as JSON
	jsonData, err := json.Marshal(cat)
	if err != nil {
		w.Debugf("Category(%s).write(): %v", cat.Name, err)
		return
	}

	// write
	ioutil.WriteFile(cat.Path, jsonData, 0666)
}

func (cat *Category) update(w *Wiki) {

	// we're probably just now creating the category, so
	// it's not gonna have any outdated information.
	if !cat.Exists() {
		return
	}

	// check each page
	now := time.Now()
	changed := false
	newPages := make(map[string]CategoryEntry, len(cat.Pages))
	for pageName, entry := range cat.Pages {

		// page no longer exists
		path := w.pathForPage(pageName)
		pageFi, err := os.Lstat(path)
		if err != nil {
			changed = true
			continue
		}

		// check if the modification date is more recent than asof date
		if entry.Asof != nil && pageFi.ModTime().After(*entry.Asof) {

			// the page has been modified since we last parsed it;
			// let's create a page that only reads variables
			// FIXME: will images, models, etc. be set?
			page := w.FindPage(pageName)
			page.VarsOnly = true

			// parse variables. if errors occur, leave as-is
			if err := page.Parse(); err != nil {
				newPages[page.Name()] = entry
				continue
			}

			// at this point, we're either removing or updating page info
			changed = true

			stillMember := false
			switch cat.Type {

			// for page links, check if the page still references the other
			case CategoryTypePage:
				_, stillMember = page.PageLinks[wikifier.CategoryNameNE(cat.Name)]

			// for images, check if the page still references the image
			case CategoryTypeImage:
				_, stillMember = page.Images[wikifier.CategoryNameNE(cat.Name)]

			// for models, check if the page still uses the model
			case CategoryTypeModel:
				_, stillMember = page.Models[wikifier.CategoryNameNE(cat.Name)]

			// for normal categories, check @category
			default:
				for _, catName := range page.Categories() {
					if catName == cat.Name {
						stillMember = true
						break
					}
				}
			}

			// page no longer belongs to the category
			if !stillMember {
				continue
			}

			// update page info
			entry.PageInfo = page.Info()
			entry.Asof = &now
		}

		newPages[pageName] = entry
	}

	// nothing changed
	if !changed {
		return
	}

	// update information
	cat.Modified = &now
	cat.Pages = newPages

	// category should be deleted
	if cat.shouldPurge(w) {
		os.Remove(cat.Path)
		return
	}

	// write update
	cat.write(w)
}

// checks if a category should be deleted
func (cat *Category) shouldPurge(w *Wiki) bool {

	// whaa? there are still pages! why you even asking?
	if len(cat.Pages) != 0 {
		return false
	}

	nameNE := wikifier.CategoryNameNE(cat.Name)
	preserve := false
	switch cat.Type {

	// note that we track references to not-yet-existent content too,
	// but if we made it to here, there are no pages referencing this

	// for page links, check if the page still exists.
	// use NewPage because it considers all possible page extensions
	case CategoryTypePage:
		preserve = w.FindPage(nameNE).Exists()

	// for images, check if the image still exists
	case CategoryTypeImage:
		_, err := os.Lstat(w.pathForImage(nameNE))
		preserve = err != nil

	// for models, check if the model still exists
	case CategoryTypeModel:
		_, err := os.Lstat(w.pathForModel(nameNE))
		preserve = err != nil

	// for normal categories, check if it's being manually preserved
	default:
		preserve = cat.Preserve

	}

	return !preserve
}

func (cat *Category) addImage(w *Wiki, imageName string, pageMaybe *wikifier.Page, dimensionsMaybe [][]int) {

	// what !!
	if cat.Type != CategoryTypeImage {
		panic("addImage() on non-image category")
	}

	cat.Preserve = true // keep until there are no more references

	// find the image dimensions if not present
	if cat.ImageInfo == nil {
		path := w.pathForImage(imageName)
		width, height := getImageDimensions(path)
		if width != 0 && height != 0 {
			cat.ImageInfo = &struct {
				Width  int `json:"width,omitempty"`
				Height int `json:"height,omitempty"`
			}{width, height}
		}
	}

	cat.addPageExtras(w, pageMaybe, dimensionsMaybe, nil)
}

// cat_check_page
func (w *Wiki) updatePageCategories(page *wikifier.Page) {

	// page metadata category
	info := page.Info()
	pageCat := w.GetSpecialCategory(page.NameNE(), CategoryTypePage)
	pageCat.PageInfo = &info
	pageCat.Preserve = true // keep until page no longer exists
	pageCat.addPageExtras(w, nil, nil, nil)

	// actual categories
	for _, name := range page.Categories() {
		w.GetCategory(name).AddPage(w, page)
	}

	// image tracking categories
	for imageName, dimensions := range page.Images {
		imageCat := w.GetSpecialCategory(imageName, CategoryTypeImage)
		imageCat.addImage(w, imageName, page, dimensions)
	}

	// page tracking categories
	for pageName, lines := range page.PageLinks {
		// note: if the page exists, the category should already exist also.
		// however, we track references to not-yet-existent pages as well
		pageCat := w.GetSpecialCategory(pageName, CategoryTypePage)
		pageCat.Preserve = true // keep until there are no more references
		pageCat.addPageExtras(w, page, nil, lines)
	}

	// model tracking categories
	for modelName := range page.Models {
		modelCat := w.GetSpecialCategory(modelName, CategoryTypeModel)
		modelCat.Preserve = true // keep until there are no more references
		modelCat.AddPage(w, page)
	}
}

// DisplayCategoryPosts returns the display result for a category.
func (w *Wiki) DisplayCategoryPosts(catName string, pageN int) interface{} {
	cat := w.GetCategory(catName)
	catName = cat.Name

	// update info
	// note: this needs to be before existence check because it may purge
	cat.update(w)

	// category does not exist
	if !cat.Exists() {
		return DisplayError{
			Error:         "Category does not exist.",
			DetailedError: "Category '" + cat.Path + "' does not exist.",
		}
	}

	// category has no pages
	// (probably shouldn't happen for normal categories, but check anyway)
	if len(cat.Pages) == 0 {
		return DisplayError{Error: "Category is empty."}
	}

	// load each page
	var pages pagesToSort
	for pageName := range cat.Pages {

		// fetch page display result
		res := w.DisplayPage(pageName)
		pageR, ok := res.(DisplayPage)
		if !ok {
			continue
		}

		// TODO: check for @category.name.main
		// and if present, set Created = infinity

		// store page result
		pages = append(pages, pageR)
	}

	// order with newest first
	sort.Sort(pages)

	// determine how many pages of pages we're gonna need
	limit := w.Opt.Category.PerPage
	numPages := 0
	if limit > 0 {
		numPages = int(math.Ceil(float64(len(pages)) / float64(limit)))
	}

	// the request is for a page beyond what we can offer
	if pageN > numPages-1 || pageN < 0 {
		return DisplayError{Error: "Page " + strconv.Itoa(pageN+1) + " does not exist."}
	}

	// if there is a limit and we exceeded it
	if limit > 0 && !(pageN == 1 && len(pages) <= limit) {
		pagesOfPages := make([]pagesToSort, numPages)

		// break down into PAGES of pages. wow.
		n := 0
		for len(pages) != 0 {

			// the length of this pageOfPages is either
			// len(pages) or limit, whichever is smaller
			thisPageLength := len(pages)
			if thisPageLength > limit {
				thisPageLength = limit
			}

			// first one on the page
			if pagesOfPages[n] == nil {
				pagesOfPages[n] = make(pagesToSort, thisPageLength, limit)
			}

			// add up to limit pages
			var i int
			for i = 0; i <= limit-1; i++ {
				if len(pages) == 0 {
					break
				}
				pagesOfPages[n][i] = pages[0] // add the next available page
				pages = pages[1:]             // remove the page just added from pending
			}

			// if that was the page we wanted, stop
			if n == pageN {
				n++
				break
			}

			n++
		}

		// only care about the page requested
		pagesOfPages = pagesOfPages[:n]
		pages = pagesOfPages[pageN]
	}

	// unfortunately we have to iterate over this 1 more time
	css := ""
	for _, page := range pages {
		if page.CSS != "" {
			css += page.CSS + "\n"
		}
	}

	return DisplayCategoryPosts{
		Pages:    pages,
		PageN:    pageN,
		NumPages: numPages,
		CSS:      css,
		Category: cat,
	}
}

// CategoryInfo represents metadata associated with a category.
type CategoryInfo struct {
	File     string     `json:"file"`               // filename
	Created  *time.Time `json:"created,omitempty"`  // creation time
	Modified *time.Time `json:"modified,omitempty"` // modify time
}

// Categories returns info about all the models in the wiki.
func (w *Wiki) Categories() []CategoryInfo {
	catNames := w.allCategoryFiles("")
	cats := make([]CategoryInfo, len(catNames))

	// cats individually
	i := 0
	for _, name := range catNames {
		cats[i] = w.CategoryInfo(name)
		i++
	}

	return cats
}

type sortableCategoryInfo CategoryInfo

func (ci sortableCategoryInfo) SortInfo() SortInfo {
	return SortInfo{
		Title:    ci.File,
		Created:  *ci.Created,
		Modified: *ci.Modified,
	}
}

// CategoriesSorted returns info about all the categories in the wiki, sorted as specified.
// Accepted sort functions are SortTitle, SortCreated, and SortModified.
func (w *Wiki) CategoriesSorted(descend bool, sorters ...SortFunc) []CategoryInfo {

	// convert to []Sortable
	cats := w.Categories()
	sorted := make([]Sortable, len(cats))
	for i, ci := range w.Categories() {
		sorted[i] = sortableCategoryInfo(ci)
	}

	// sort
	var sorter sort.Interface = sorter(sorted, sorters...)
	if descend {
		sorter = sort.Reverse(sorter)
	}
	sort.Sort(sorter)

	// convert back to []CategoryInfo
	for i, si := range sorted {
		cats[i] = CategoryInfo(si.(sortableCategoryInfo))
	}

	return cats
}

// CategoryMap returns a map of model name to CategoryInfo for all models in the wiki.
func (w *Wiki) CategoryMap() map[string]CategoryInfo {
	catNames := w.allCategoryFiles("")
	cats := make(map[string]CategoryInfo, len(catNames))

	// models individually
	for _, name := range catNames {
		cats[name] = w.CategoryInfo(name)
	}

	return cats
}

// CategoryInfo is an inexpensive request for info on a category.
func (w *Wiki) CategoryInfo(name string) (info CategoryInfo) {
	cat := w.GetCategory(name)

	// this stuff is available to all
	info.File = cat.File
	info.Modified = cat.Modified
	info.Created = cat.Created

	return
}

// logic for sorting pages by time

type pagesToSort []DisplayPage

func (p pagesToSort) Len() int {
	return len(p)
}

func (p pagesToSort) Less(i, j int) bool {

	// neither have time set; fall back to alphabetical
	if p[i].Created == nil && p[j].Created == nil {
		names := sort.StringSlice([]string{p[i].Name, p[j].Name})
		return names.Less(0, 1)
	}

	// one has no time set
	if p[j].Created == nil {
		return true
	}
	if p[i].Created == nil {
		return false
	}

	return p[i].Created.After(*p[j].Created)
}

func (p pagesToSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
