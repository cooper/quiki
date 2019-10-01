package wiki

import (
	"time"

	"github.com/cooper/quiki/wikifier"
)

// CategoryType describes the type of a Category.
type CategoryType string

var (
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

	// category filename, including the .cat extension
	File string

	// category name without extension
	Name string

	// human-readable category title
	Title string

	// time when the category was created
	Created time.Time

	// time when the category was last modified.
	// this is updated when pages are added and deleted
	Modified time.Time

	// pages in the category. keys are filenames
	Pages map[string]CategoryEntry

	// when true, the category is preserved even when no pages remain
	Preserve bool

	// EXTRAS

	// if applicable, this is the type of pseudocategory.
	// for normal categories, this is empty
	Type CategoryType

	// for CategoryTypePage, this is the info for the tracked page
	PageInfo wikifier.PageInfo
}

// A CategoryEntry describes a page that belongs to a category.
type CategoryEntry struct {

	// time at which the page metadata in this category file was last updated.
	// this is compared against page file modification time
	Asof time.Time

	// embedded page info
	// note this info is accurate only as of the Asof time
	wikifier.PageInfo

	// EXTRAS

	// for CategoryTypeImage, an array of image dimensions used on this page.
	// dimensions are guaranteed to be positive integers. the number of elements will
	// always be even, since each occurence of the image produces two (width and then height)
	Dimensions [][]int

	// for CategoryTypePage, an array of line numbers on which the tracked page is
	// referenced on the page described by this entry
	Lines []int
}
