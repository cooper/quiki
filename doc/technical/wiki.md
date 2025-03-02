# wiki
--
    import "."


## Usage

```go
const (
	// CategoryTypeImage is a type of category that tracks which pages use an image.
	CategoryTypeImage CategoryType = "image"

	// CategoryTypeModel is a metacategory that tracks which pages use a model.
	CategoryTypeModel = "model"

	// CategoryTypePage is a metacategory that tracks which pages reference another page.
	CategoryTypePage = "page"
)
```

#### func  AvailableBaseWikis

```go
func AvailableBaseWikis() []string
```
AvailableBaseWikis returns a list of available embedded base wikis.

#### func  CreateWiki

```go
func CreateWiki(path, basePath string, opts CreateWikiOpts) error
```
CreateWiki creates a new wiki at the specified path using a base wiki directory.

#### func  CreateWikiFS

```go
func CreateWikiFS(path string, fsys fs.FS, opts CreateWikiOpts) error
```
CreateWiki creates a new wiki at the specified path using a base wiki fs.

#### func  CreateWikiFromResource

```go
func CreateWikiFromResource(path, resourceName string, opts CreateWikiOpts) error
```
CreateWikiFromResource creates a new wiki at the specified path using a base
wiki resource.

#### func  SortAuthor

```go
func SortAuthor(p, q Sortable) bool
```
SortAuthor is a SortFunc for sorting items alphabetically by author.

#### func  SortCreated

```go
func SortCreated(p, q Sortable) bool
```
SortCreated is a SortFunc for sorting items by creation time.

#### func  SortDimensions

```go
func SortDimensions(p, q Sortable) bool
```
SortDimensions is a SortFunc for sorting images by their dimensions.

#### func  SortModified

```go
func SortModified(p, q Sortable) bool
```
SortModified is a SortFunc for sorting items by modification time.

#### func  SortTitle

```go
func SortTitle(p, q Sortable) bool
```
SortTitle is a SortFunc for sorting items alphabetically by title.

#### func  ValidBranchName

```go
func ValidBranchName(name string) bool
```
ValidBranchName returns whether a branch name is valid.

quiki branch names may contain word-like characters `\w` and forward slash (`/`)
but may not start or end with a slash.

#### type Category

```go
type Category struct {

	// category path
	Path string `json:"-"`

	// category filename, including the .cat extension
	File string `json:"file"`

	// category name without extension
	Name   string `json:"name,omitempty"`
	FileNE string `json:"file_ne,omitempty"` // alias for consistency

	// human-readable category title
	Title string `json:"title,omitempty"`

	// number of posts per page when displaying as posts
	// (@category.per_page)
	PerPage int `json:"per_page,omitempty"`

	// time when the category was created
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

	// time when the category was last modified.
	// this is updated when pages are added and deleted
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted

	// time when the category metafile was last read.
	Asof *time.Time `json:"asof,omitempty"`

	// pages in the category. keys are filenames
	Pages map[string]CategoryEntry `json:"pages,omitempty"`

	// when true, the category is preserved even when no pages remain
	Preserve bool `json:"preserve,omitempty"`

	// if applicable, this is the type of special category.
	// for normal categories, this is empty
	Type CategoryType `json:"type,omitempty"`

	// for CategoryTypePage, this is the info for the tracked page
	PageInfo *wikifier.PageInfo `json:"page_info,omitempty"`

	// for CategoryTypeModel, this is the info for the tracked model
	ModelInfo *wikifier.ModelInfo `json:"model_info,omitempty"`

	// for CategoryTypeImage, this is the info for the tracked image
	ImageInfo *struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"image_info,omitempty"`
}
```

A Category is a collection of pages pertaining to a topic.

A page can belong to many categories. Category memberships and metadata are
stored in JSON manifests.

#### func (*Category) AddPage

```go
func (cat *Category) AddPage(w *Wiki, page *wikifier.Page)
```
AddPage adds a page to a category.

If the page already belongs and any information has changed, the category is
updated. If force is true,

#### func (*Category) Exists

```go
func (cat *Category) Exists() bool
```
Exists returns whether a category currently exists.

#### type CategoryEntry

```go
type CategoryEntry struct {

	// time at which the page metadata in this category file was last updated.
	// this is compared against page file modification time
	Asof *time.Time `json:"asof,omitempty"`

	// embedded page info
	// note this info is accurate only as of the Asof time
	wikifier.PageInfo

	// for CategoryTypeImage, an array of image dimensions used on this page.
	// dimensions are guaranteed to be positive integers. the number of elements will
	// always be even, since each occurrence of the image produces two (width and then height)
	Dimensions [][]int `json:"dimensions,omitempty"`

	// for CategoryTypePage, an array of line numbers on which the tracked page is
	// referenced on the page described by this entry
	Lines []int `json:"lines,omitempty"`
}
```

A CategoryEntry describes a page that belongs to a category.

#### type CategoryInfo

```go
type CategoryInfo struct {
	*Category
}
```

CategoryInfo represents metadata associated with a category.

#### type CategoryType

```go
type CategoryType string
```

CategoryType describes the type of a Category.

#### type CommitOpts

```go
type CommitOpts struct {

	// Comment is the commit description.
	Comment string

	// Name is the fullname of the user committing changes.
	Name string

	// Email is the email address of the user committing changes.
	Email string

	// Time is the timestamp to associate with the revision.
	// If unspecified, current time is used.
	Time time.Time
}
```

CommitOpts describes the options for a wiki revision.

#### type CreateWikiOpts

```go
type CreateWikiOpts struct {
	WikiName     string
	TemplateName string
	MainPage     string
	ErrorPage    string
}
```


#### type DisplayCategoryPosts

```go
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
```

DisplayCategoryPosts represents a category result to display.

#### type DisplayError

```go
type DisplayError struct {
	// a human-readable error string. sensitive info is never
	// included, so this may be shown to users
	Error string

	// a more detailed human-readable error string that MAY contain
	// sensitive data. can be used for debugging and logging but should
	// not be presented to users
	DetailedError string

	// HTTP status code. if zero, 404 should be used
	Status int

	// if the error occurred during parsing, this is the position.
	// for all non-parsing errors, this is 0:0
	Pos wikifier.Position

	// true if the content cannot be displayed because it has
	// not yet been published for public access
	Draft bool
}
```

DisplayError represents an error result to display.

#### func (DisplayError) ErrorAsWarning

```go
func (e DisplayError) ErrorAsWarning() wikifier.Warning
```

#### type DisplayFile

```go
type DisplayFile struct {

	// file name relative to wiki root.
	// path delimiter '/' is always used, regardless of OS.
	File string `json:"file,omitempty"`

	// absolute file path of the file.
	// OS-specific path delimiter is used.
	Path string `json:"path,omitempty"`

	// the plain text file content
	Content string `json:"-"`

	// time when the file was last modified
	Modified *time.Time `json:"modified,omitempty"`

	// for pages/models/etc, parser warnings and error
	Warnings []wikifier.Warning `json:"parse_warnings,omitempty"`
	Error    *wikifier.Warning  `json:"parse_error,omitempty"`
}
```

DisplayFile represents a plain text file to display.

#### type DisplayImage

```go
type DisplayImage struct {

	// basename of the scaled image file
	File string `json:"file,omitempty"`

	// absolute path to the scaled image.
	// this file should be served to the user
	Path string `json:"path,omitempty"`

	// absolute path to the full-size image.
	// if the full-size image is being displayed, same as Path
	FullsizePath string `json:"fullsize_path,omitempty"`

	// image type
	// 'png' or 'jpeg'
	ImageType string `json:"image_type,omitempty"`

	// mime 'image/png' or 'image/jpeg'
	// suitable for the Content-Type header
	Mime string `json:"mime,omitempty"`

	// bytelength of image data
	// suitable for use in the Content-Length header
	Length int64 `json:"length,omitempty"`

	// time when the image was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the image file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP format for Last-Modified

	// true if the content being sered was read from a cache file.
	// opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if the content generated in order to fulfill this request was
	// written to cache. this can only been true when Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`
}
```

DisplayImage represents an image to display.

#### type DisplayPage

```go
type DisplayPage struct {

	// basename of the page, with the extension
	File string `json:"file,omitempty"`

	// basename of the page, without the extension
	Name string `json:"name,omitempty"`

	// absolute file path of the page
	Path string `json:"path,omitempty"`

	// the page content (HTML)
	Content wikifier.HTML `json:"-"`

	// time when the page was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the page file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted for Last-Modified

	// CSS generated for the page from style{} blocks
	CSS string `json:"css,omitempty"`

	// true if this content was read from a cache file. opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated on the fly.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if this request resulted in the writing of a new cache file.
	// this can only be true if Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`

	// true if this request resulted in the writing of a text file.
	// this can only be true if Generated is true
	TextGenerated bool `json:"text_gen,omitempty"`

	// true if the page has not yet been published for public viewing.
	// this only occurs when it is specified that serving drafts is OK,
	// since normally a draft page instead results in a DisplayError.
	Draft bool `json:"draft,omitempty"`

	// warnings and errors produced by the parser
	Warnings []wikifier.Warning `json:"warnings,omitempty"`

	// time when the page was created, as extracted from
	// the special @page.created variable
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

	// name of the page author, as extracted from the special @page.author
	// variable
	Author string `json:"author,omitempty"`

	// list of categories the page belongs to, without the '.cat' extension
	Categories []string `json:"categories,omitempty"`

	// page title as extracted from the special @page.title variable, including
	// any possible HTML-encoded formatting
	FmtTitle wikifier.HTML `json:"fmt_title,omitempty"`

	// like FmtTitle except that all text formatting has been stripped.
	// suitable for use in the <title> tag
	Title string `json:"title,omitempty"`

	// page description as extracted from the special @page.desc variable.
	Description string `json:"desc,omitempty"`

	// page keywords as extracted from the special @page.keywords variable
	Keywords []string `json:"keywords,omitempty"`

	// first formatting-stripped 25 words of page, up to 150 chars
	Preview string `json:"preview,omitempty"`
}
```

DisplayPage represents a page result to display.

#### type DisplayRedirect

```go
type DisplayRedirect struct {

	// a relative or absolute URL to which the request should redirect,
	// suitable for use in a Location header
	Redirect string
}
```

DisplayRedirect represents a page redirect to follow.

#### type ImageInfo

```go
type ImageInfo struct {
	File       string     `json:"file"`               // filename
	Width      int        `json:"width,omitempty"`    // full-size width
	Height     int        `json:"height,omitempty"`   // full-size height
	Created    *time.Time `json:"created,omitempty"`  // creation time
	Modified   *time.Time `json:"modified,omitempty"` // modify time
	Dimensions [][]int    `json:"-"`                  // dimensions used throughout the wiki
}
```

ImageInfo represents a full-size image on the wiki.

#### type RevisionInfo

```go
type RevisionInfo struct {
	Id      string    `json:"id"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
```

RevisionInfo contains information about a specific revision.

#### type SizedImage

```go
type SizedImage struct {
	// for example mydir/100x200-myimage@3x.png
	Width, Height int    // 100, 200 (dimensions as requested)
	Scale         int    // 3 (scale as requested)
	Prefix        string // mydir
	RelNameNE     string // myimage (name without extension)
	Ext           string // png (extension)
}
```

SizedImage represents an image in specific dimensions.

#### func  SizedImageFromName

```go
func SizedImageFromName(name string) SizedImage
```
SizedImageFromName returns a SizedImage given an image name.

#### func (SizedImage) FullSizeName

```go
func (img SizedImage) FullSizeName() string
```
FullSizeName returns the name of the full-size image.

#### func (SizedImage) ScaleName

```go
func (img SizedImage) ScaleName() string
```
ScaleName returns the image name with dimensions and scale.

#### func (SizedImage) TrueHeight

```go
func (img SizedImage) TrueHeight() int
```
TrueHeight returns the actual image height when the Scale is taken into
consideration.

#### func (SizedImage) TrueName

```go
func (img SizedImage) TrueName() string
```
TrueName returns the image name with true dimensions.

#### func (SizedImage) TrueNameNE

```go
func (img SizedImage) TrueNameNE() string
```
TrueNameNE is like TrueName but without the extension.

#### func (SizedImage) TrueWidth

```go
func (img SizedImage) TrueWidth() int
```
TrueWidth returns the actual image width when the Scale is taken into
consideration.

#### type SortFunc

```go
type SortFunc func(p, q Sortable) bool
```

SortFunc is a type for functions that can sort items.

#### type SortInfo

```go
type SortInfo struct {
	Title      string
	Author     string
	Created    time.Time
	Modified   time.Time
	Dimensions []int
}
```

SortInfo is the data returned from Sortable items for sorting wiki resources.

#### type Sortable

```go
type Sortable interface {
	SortInfo() SortInfo
}
```

Sortable is the interface that allows quiki to sort wiki resources.

#### type Wiki

```go
type Wiki struct {
	ConfigFile string
	Opt        wikifier.PageOpt
	Auth       *authenticator.Authenticator
}
```

A Wiki represents a quiki website.

#### func  NewWiki

```go
func NewWiki(path string) (*Wiki, error)
```
NewWiki creates a Wiki given its directory path.

#### func (*Wiki) AbsFilePath

```go
func (w *Wiki) AbsFilePath(relPath string) string
```
AbsFilePath takes a relative path to a file within the wiki (e.g.
`pages/mypage.page`), joins it with the wiki directory, and evaluates it with
`filepath.Abs()`. The result is an absolute path which may or may not exist.

If the file is a symlink, it is followed. Thus, it is possible for the resulting
path to exist outside the wiki directory. If that is not desired, use
unresolvedAbsFilePath instead.

#### func (*Wiki) Branch

```go
func (w *Wiki) Branch(name string) (*Wiki, error)
```
Branch returns a Wiki instance for this wiki at another branch. If the branch
does not exist, an error is returned.

#### func (*Wiki) BranchNames

```go
func (w *Wiki) BranchNames() ([]string, error)
```
BranchNames returns the revision branches available.

#### func (*Wiki) Categories

```go
func (w *Wiki) Categories() []CategoryInfo
```
Categories returns info about all the models in the wiki.

#### func (*Wiki) CategoriesSorted

```go
func (w *Wiki) CategoriesSorted(descend bool, sorters ...SortFunc) []CategoryInfo
```
CategoriesSorted returns info about all the categories in the wiki, sorted as
specified. Accepted sort functions are SortTitle, SortCreated, and SortModified.

#### func (*Wiki) CategoryInfo

```go
func (w *Wiki) CategoryInfo(name string) (info CategoryInfo)
```
CategoryInfo is an inexpensive request for info on a category.

#### func (*Wiki) CategoryMap

```go
func (w *Wiki) CategoryMap() map[string]CategoryInfo
```
CategoryMap returns a map of model name to CategoryInfo for all models in the
wiki.

#### func (*Wiki) CreateModel

```go
func (w *Wiki) CreateModel(title string, content []byte, commit CommitOpts) (string, error)
```
CreateModel creates a new model file.

#### func (*Wiki) CreatePage

```go
func (w *Wiki) CreatePage(where string, title string, content []byte, commit CommitOpts) (string, error)
```
CreatePage creates a new page file. If content is empty, a default page is
created.

#### func (*Wiki) CreatePageFolder

```go
func (w *Wiki) CreatePageFolder(where string, name string) error
```
CreatePageFolder creates a new page folder.

#### func (*Wiki) Debug

```go
func (w *Wiki) Debug(i ...any)
```
Debug logs debug info for a wiki.

#### func (*Wiki) Debugf

```go
func (w *Wiki) Debugf(format string, i ...any)
```
Debugf logs debug info for a wiki.

#### func (*Wiki) DeleteFile

```go
func (w *Wiki) DeleteFile(name string, commit CommitOpts) error
```
DeleteFile deletes a file in the wiki.

The filename must be relative to the wiki directory. If the file does not exist,
an error is returned. If the file exists and is a symbolic link, the link itself
is deleted, not the target file.

This is a low-level API that allows deleting any file within the wiki directory,
so it should not be utilized directly by frontends. Use DeletePage, DeleteModel,
or DeleteImage instead.

#### func (*Wiki) Dir

```go
func (w *Wiki) Dir(dirs ...string) string
```
Dir returns the absolute path to the resolved wiki directory. If the wiki
directory is a symlink, it is followed.

Optional path components can be passed as arguments to be joined with the wiki
root by the path separator.

#### func (*Wiki) DisplayCategoryPosts

```go
func (w *Wiki) DisplayCategoryPosts(catName string, pageN int) any
```
DisplayCategoryPosts returns the display result for a category.

#### func (*Wiki) DisplayFile

```go
func (w *Wiki) DisplayFile(path string) any
```
DisplayFile returns the display result for a plain text file.

#### func (*Wiki) DisplayImage

```go
func (w *Wiki) DisplayImage(name string) any
```
DisplayImage returns the display result for an image.

#### func (*Wiki) DisplayPage

```go
func (w *Wiki) DisplayPage(name string) any
```
DisplayPage returns the display result for a page.

#### func (*Wiki) DisplayPageDraft

```go
func (w *Wiki) DisplayPageDraft(name string, draftOK bool) any
```
DisplayPageDraft returns the display result for a page.

Unlike DisplayPage, if draftOK is true, the content is served even if it is
marked as draft.

#### func (*Wiki) DisplaySizedImage

```go
func (w *Wiki) DisplaySizedImage(img SizedImage) any
```
DisplaySizedImage returns the display result for an image in specific
dimensions.

#### func (*Wiki) DisplaySizedImageGenerate

```go
func (w *Wiki) DisplaySizedImageGenerate(img SizedImage, generateOK bool) any
```
DisplaySizedImageGenerate returns the display result for an image in specific
dimensions and allows images to be generated in any dimension.

#### func (*Wiki) FindPage

```go
func (w *Wiki) FindPage(name string) (p *wikifier.Page)
```
FindPage attempts to find a page on this wiki given its name, regardless of the
file format or filename case.

If a page by this name exists, the returned page represents it. Otherwise, a new
page representing the lowercased, normalized .page file is returned in the
standard quiki filename format.

#### func (*Wiki) GetCategory

```go
func (w *Wiki) GetCategory(name string) *Category
```
GetCategory loads or creates a category.

#### func (*Wiki) GetLatestCommitHash

```go
func (w *Wiki) GetLatestCommitHash() (string, error)
```
GetLatestCommitHash returns the most recent commit hash.

#### func (*Wiki) GetSpecialCategory

```go
func (w *Wiki) GetSpecialCategory(name string, typ CategoryType) *Category
```
GetSpecialCategory loads or creates a special category given the type.

#### func (*Wiki) ImageInfo

```go
func (w *Wiki) ImageInfo(name string) (info ImageInfo)
```
ImageInfo returns info for an image given its full-size name.

#### func (*Wiki) ImageMap

```go
func (w *Wiki) ImageMap() map[string]ImageInfo
```
ImageMap returns a map of image filename to ImageInfo for all images in the
wiki.

#### func (*Wiki) Images

```go
func (w *Wiki) Images() []ImageInfo
```
Images returns info about all the images in the wiki.

#### func (*Wiki) ImagesSorted

```go
func (w *Wiki) ImagesSorted(descend bool, sorters ...SortFunc) []ImageInfo
```
ImagesSorted returns info about all the pages in the wiki, sorted as specified.
Accepted sort functions are SortTitle, SortAuthor, SortCreated, SortModified,
and SortDimensions.

#### func (*Wiki) Log

```go
func (w *Wiki) Log(i ...any)
```
Log logs info for a wiki.

#### func (*Wiki) Logf

```go
func (w *Wiki) Logf(format string, i ...any)
```
Logf logs info for a wiki.

#### func (*Wiki) ModelInfo

```go
func (w *Wiki) ModelInfo(name string) (info wikifier.ModelInfo)
```
ModelInfo is an inexpensive request for info on a model. It uses cached metadata
rather than generating the model and extracting variables.

#### func (*Wiki) ModelMap

```go
func (w *Wiki) ModelMap() map[string]wikifier.ModelInfo
```
ModelMap returns a map of model name to wikifier.ModelInfo for all models in the
wiki.

#### func (*Wiki) Models

```go
func (w *Wiki) Models() []wikifier.ModelInfo
```
Models returns info about all the models in the wiki.

#### func (*Wiki) ModelsSorted

```go
func (w *Wiki) ModelsSorted(descend bool, sorters ...SortFunc) []wikifier.ModelInfo
```
ModelsSorted returns info about all the models in the wiki, sorted as specified.
Accepted sort functions are SortTitle, SortAuthor, SortCreated, and
SortModified.

#### func (*Wiki) NewBranch

```go
func (w *Wiki) NewBranch(name string) (*Wiki, error)
```
NewBranch is like Branch, except it creates the branch at the current master
revision if it does not yet exist.

#### func (*Wiki) PageInfo

```go
func (w *Wiki) PageInfo(name string) (info wikifier.PageInfo)
```
PageInfo is an inexpensive request for info on a page. It uses cached metadata
rather than generating the page and extracting variables.

#### func (*Wiki) PageMap

```go
func (w *Wiki) PageMap() map[string]wikifier.PageInfo
```
PageMap returns a map of page name to PageInfo for all pages in the wiki.

#### func (*Wiki) Pages

```go
func (w *Wiki) Pages() []wikifier.PageInfo
```
Pages returns info about all the pages in the wiki.

#### func (*Wiki) PagesAndDirs

```go
func (w *Wiki) PagesAndDirs(where string) ([]wikifier.PageInfo, []string)
```
PagesAndDirs returns info about all the pages and directories in a directory.

#### func (*Wiki) PagesAndDirsSorted

```go
func (w *Wiki) PagesAndDirsSorted(where string, descend bool, sorters ...SortFunc) ([]wikifier.PageInfo, []string)
```
PagesAndDirsSorted returns info about all the pages and directories in a
directory, sorted as specified. Accepted sort functions are SortTitle,
SortAuthor, SortCreated, and SortModified. Directories are always sorted
alphabetically (but still respect the descend flag).

#### func (*Wiki) PagesInDir

```go
func (w *Wiki) PagesInDir(where string) []wikifier.PageInfo
```
PagesInDir returns info about all the pages in the specified directory.

#### func (*Wiki) PagesSorted

```go
func (w *Wiki) PagesSorted(descend bool, sorters ...SortFunc) []wikifier.PageInfo
```
PagesSorted returns info about all the pages in the wiki, sorted as specified.
Accepted sort functions are SortTitle, SortAuthor, SortCreated, and
SortModified.

#### func (*Wiki) PathForCategory

```go
func (w *Wiki) PathForCategory(catName string, createOK bool) string
```
PathForCategory returns the absolute path for a category. If createOK is true,
it creates directories for the path components that do not exist.

#### func (*Wiki) PathForImage

```go
func (w *Wiki) PathForImage(imageName string) string
```
PathForImage returns the absolute path for an image.

#### func (*Wiki) PathForMetaCategory

```go
func (w *Wiki) PathForMetaCategory(catName string, catType CategoryType, createOK bool) string
```
PathForMetaCategory returns the absolute path for a meta category. Meta
categories are used for internal categorization and not exposed in the wiki.

If createOK is true, it creates directories for the path components that do not
exist.

#### func (*Wiki) PathForModel

```go
func (w *Wiki) PathForModel(modelName string) string
```
PathForModel returns the absolute path for a model.

#### func (*Wiki) PathForPage

```go
func (w *Wiki) PathForPage(pageName string) string
```
PathForPage returns the absolute path for a page.

#### func (*Wiki) Pregenerate

```go
func (w *Wiki) Pregenerate() (results []any)
```
Pregenerate simulates requests for all wiki resources such that content caches
can be pregenerated and stored.

#### func (*Wiki) RelPath

```go
func (w *Wiki) RelPath(absPath string) string
```
RelPath takes an absolute file path and attempts to make it relative to the wiki
directory, regardless of whether the path exists.

If the path can be made relative without following symlinks, this is preferred.
If that fails, symlinks in absPath are followed and a second attempt is made.

In any case the path cannot be made relative to the wiki directory, an empty
string is returned.

#### func (*Wiki) RevisionsMatchingPage

```go
func (w *Wiki) RevisionsMatchingPage(nameOrPath string) ([]RevisionInfo, error)
```
RevisionsMatchingPage returns a list of commit infos matching a page file.

#### func (*Wiki) UnresolvedAbsFilePath

```go
func (w *Wiki) UnresolvedAbsFilePath(relPath string) string
```
UnresolvedAbsFilePath takes a relative path to a file within the wiki (e.g.
`pages/mypage.page`) and joins it with the absolute path to the wiki directory.
The result is an absolute path which may or may not exist.

Symlinks are not followed. If that is desired, use absoluteFilePath instead.

#### func (*Wiki) WriteConfig

```go
func (w *Wiki) WriteConfig(content []byte, commit CommitOpts) error
```
WriteConfig writes the wiki configuration file.

#### func (*Wiki) WriteImage

```go
func (w *Wiki) WriteImage(name string, content []byte, createOK bool, commit CommitOpts) error
```
WriteImage writes an image file.

#### func (*Wiki) WriteModel

```go
func (w *Wiki) WriteModel(name string, content []byte, createOK bool, commit CommitOpts) error
```
WriteModel writes a model file.

#### func (*Wiki) WritePage

```go
func (w *Wiki) WritePage(name string, content []byte, createOK bool, commit CommitOpts) error
```
WritePage writes a page file.
