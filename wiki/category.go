package wiki

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
// A page can belong to many categories. Category memberships and metadta
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
	// always be even, since each occurence of the image produces two (width and then height)
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
	name = wikifier.CategoryNameNE(name, false)
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
		log.Printf("GetCategory(%s): %v", name, err)
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
func (cat *Category) AddPage(page *wikifier.Page) {
	cat.addPageExtras(page, nil, nil)
}

func (cat *Category) addPageExtras(pageMaybe *wikifier.Page, dimensions [][]int, lines []int) {
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
	cat.write()
}

// Exists returns whether a category currently exists.
func (cat *Category) Exists() bool {
	_, err := os.Lstat(cat.Path)
	return err == nil
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
	info := page.Info()
	pageCat := w.GetSpecialCategory(page.Name(), CategoryTypePage)
	pageCat.PageInfo = &info
	pageCat.Preserve = true // keep until page no longer exists
	pageCat.addPageExtras(nil, nil, nil)

	// actual categories
	for _, name := range page.Categories() {
		w.GetCategory(name).AddPage(page)
	}

	// image tracking categories
	for imageName, dimensions := range page.Images {
		imageCat := w.GetSpecialCategory(imageName, CategoryTypeImage)
		imageCat.Preserve = true // keep until there are no more references

		// find the image dimensions if not present
		if imageCat.ImageInfo == nil {
			path := w.pathForImage(imageName)
			w, h := getImageDimensions(path)
			if w != 0 && h != 0 {
				imageCat.ImageInfo = &struct {
					Width  int `json:"width,omitempty"`
					Height int `json:"height,omitempty"`
				}{w, h}
			}
		}

		imageCat.addPageExtras(page, dimensions, nil)
	}

	// page tracking categories
	for pageName, lines := range page.PageLinks {
		// note: if the page exists, the category should already exist also.
		// however, we track references to not-yet-existent pages as well
		pageCat := w.GetSpecialCategory(pageName, CategoryTypePage)
		pageCat.Preserve = true // keep until there are no more references
		pageCat.addPageExtras(page, nil, lines)
	}

	// TODO: model categories
}

// DisplayCategoryPosts returns the display result for a category.
func (w *Wiki) DisplayCategoryPosts(catName string, pageN int) interface{} {
	cat := w.GetCategory(catName)
	catName = cat.Name

	// my ($wiki, $cat_name, %opts) = @_; my $result = {};
	// $cat_name = cat_name($cat_name);
	// my $cat_name_ne = cat_name_ne($cat_name);
	// my ($err, $pages, $title) = $wiki->cat_get_pages($cat_name,
	// 	cat_type => $opts{cat_type}
	// );

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
		return DisplayError{
			Error: "Category is empty.",
		}
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
		// and if present, set CreatedUnix = infinity

		// store page result
		pages = append(pages, pageR)
	}

	// order with newest first
	sort.Sort(pages)

	// determine how many pages of pages we're gonna need
	limit := w.Opt.Category.PerPage
	numPages := 0
	if limit > 0 {
		numPages = len(pages)/limit + 1
	}

	// the request is for a page beyond what we can offer
	if pageN > numPages-1 || pageN < 0 {
		return DisplayError{Error: "Page " + strconv.Itoa(pageN+1) + " does not exist."}
	}

	// if there is a limit and we exceeded it
	if limit > 0 && !(pageN == 1 && len(pages) <= limit) {
		pagesOfPages := make([]pagesToSort, 0, numPages)

		// break down into PAGES of pages. wow.
		n := 0
		for len(pages) != 0 {

			// first one on the page
			var thisPage pagesToSort
			if n < len(pagesOfPages) {
				thisPage = pagesOfPages[n]
			} else {
				thisPage = make(pagesToSort, limit)
				pagesOfPages = pagesOfPages[:n+1]
				pagesOfPages[n] = thisPage
			}

			// add up to limit pages
			var i int
			for i = 0; i <= limit-1; i++ {
				if len(pages) == 0 {
					break
				}
				thisPage[i] = pages[0]
				pages = pages[1:]
			}
			thisPage = thisPage[:i]

			// if that was the page we wanted, stop
			if n == pageN {
				break
			}

			n++
		}

		// only care about the page requested
		pages = pagesOfPages[pageN]
	}

	css := ""

	return DisplayCategoryPosts{
		Pages:    pages,
		PageN:    pageN,
		NumPages: numPages,
		CSS:      css,
		Category: cat,
	}
}

// logic for sorting pages by time

type pagesToSort []DisplayPage

func (p pagesToSort) Len() int {
	return len(p)
}

func (p pagesToSort) Less(i, j int) bool {
	if p[j].Created == nil {
		return true
	}
	if p[i].Created == nil {
		return false
	}
	return p[i].Created.Before(*p[j].Created)
}

func (p pagesToSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
