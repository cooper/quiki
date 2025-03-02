# wikifier
--
    import "."


## Usage

```go
const (
	// PageOptExternalTypeQuiki is the external type for quiki sites.
	PageOptExternalTypeQuiki PageOptExternalType = "quiki"

	// PageOptExternalTypeMediaWiki is the external type for MediaWiki sites.
	PageOptExternalTypeMediaWiki = "mediawiki"

	// PageOptExternalTypeNone is an external type that can be used for websites that
	// perform no normalization of page targets beyond normal URI escaping.
	PageOptExternalTypeNone = "none"
)
```

#### func  CategoryName

```go
func CategoryName(name string) string
```
CategoryName returns a clean category name.

#### func  CategoryNameNE

```go
func CategoryNameNE(name string) string
```
CategoryNameNE returns a clean category with No Extension.

#### func  InjectPageOpt

```go
func InjectPageOpt(page *Page, opt *PageOpt) error
```
InjectPageOpt extracts page options found in the specified page and injects them
into the provided PageOpt pointer.

#### func  MakeDir

```go
func MakeDir(dir, name string)
```
MakeDir creates directories recursively.

#### func  ModelName

```go
func ModelName(name string) string
```
ModelName returns a clean model name.

#### func  PageName

```go
func PageName(name string) string
```
PageName returns a clean page name.

#### func  PageNameExt

```go
func PageNameExt(name, ext string) string
```
PageNameExt returns a clean page name with the provided extension.

#### func  PageNameLink

```go
func PageNameLink(name string) string
```
PageNameLink returns a clean page name without the extension.

#### func  PageNameNE

```go
func PageNameNE(name string) string
```
PageNameNE returns a clean page name with No Extension.

#### func  ScaleString

```go
func ScaleString(name string, retina []int) string
```
ScaleString returns a string of scaled image names for use in srcset.

#### func  UniqueFilesInDir

```go
func UniqueFilesInDir(dir string, extensions []string, thisDirOnly bool) ([]string, error)
```
UniqueFilesInDir recursively scans a directory for files matching the requested
extensions, resolves symlinks, and returns a list of unique files. That is, if
more than one link resolves to the same thing (as is the case for quiki page
redirects), there is only one occurrence in the output.

#### type AttributedObject

```go
type AttributedObject interface {

	// getters
	Get(key string) (any, error)
	GetBool(key string) (bool, error)
	GetStr(key string) (string, error)
	GetBlock(key string) (block, error)
	GetObj(key string) (AttributedObject, error)

	// setters
	Set(key string, value any) error
	Unset(key string) error
	// contains filtered or unexported methods
}
```

An AttributedObject is any object on which you can set and retrieve attributes.

For example, a Page is an attributed object since it contains variables.
Likewise, a Map is an attributed object because it has named properties.

#### type FmtOpt

```go
type FmtOpt struct {
	Pos        Position // position used for warnings (set internally)
	NoEntities bool     // disables html entity conversion
	NoWarnings bool     // silence warnings for undefined variables
}
```

FmtOpt describes options for page.FmtOpts.

#### type HTML

```go
type HTML string
```

HTML encapsulates a string to indicate that it is preformatted HTML. It lets
quiki's parsers know not to attempt to format it any further.

#### type List

```go
type List struct {
}
```

List represents a list of items. It is a quiki data type as well as the base of
many block types.

#### func  NewList

```go
func NewList(mb block) *List
```
NewList creates a new list, given the main block of the page it is to be
associated with.

#### func (List) Get

```go
func (scope List) Get(key string) (any, error)
```
Get fetches a a value regardless of type.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to read a property of an object that does not support properties,
such as a string, Get returns an error.

If the key is valid but nothing exists at it, Get returns (nil, nil).

#### func (List) GetBlock

```go
func (scope List) GetBlock(key string) (block, error)
```
GetBlock is like Get except it always returns a block.

#### func (List) GetBool

```go
func (scope List) GetBool(key string) (bool, error)
```
GetBool is like Get except it always returns a boolean.

#### func (List) GetObj

```go
func (scope List) GetObj(key string) (AttributedObject, error)
```
GetObj is like Get except it always returns an AttributedObject.

#### func (List) GetStr

```go
func (scope List) GetStr(key string) (string, error)
```
GetStr is like Get except it always returns a string.

If the value is HTML, it is converted to a string.

#### func (List) GetStrList

```go
func (scope List) GetStrList(key string) ([]string, error)
```
GetStrList is like Get except it always returns a list of strings.

If the value is a `list{}` block, the list's values are returned, with
non-strings quietly filtered out.

If the value is a string, it is treated as a comma-separated list, and each item
is trimmed of prepending or suffixing whitespace.

#### func (List) Set

```go
func (scope List) Set(key string, value any) error
```
Set sets a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to write to a property of an object that does not support
properties, such as a string, Set returns an error.

#### func (List) String

```go
func (b List) String() string
```

#### func (List) Unset

```go
func (scope List) Unset(key string) error
```
Unset removes a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to unset a property of an object that does not support properties,
such as a string, Unset returns an error.

#### type Map

```go
type Map struct {
}
```

Map represents a Key-value dictionary. It is a quiki data type as well as the
base of many block types.

#### func  NewMap

```go
func NewMap(mb block) *Map
```
NewMap creates a new map, given the main block of the page it is to be
associated with.

#### func (Map) Get

```go
func (scope Map) Get(key string) (any, error)
```
Get fetches a a value regardless of type.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to read a property of an object that does not support properties,
such as a string, Get returns an error.

If the key is valid but nothing exists at it, Get returns (nil, nil).

#### func (Map) GetBlock

```go
func (scope Map) GetBlock(key string) (block, error)
```
GetBlock is like Get except it always returns a block.

#### func (Map) GetBool

```go
func (scope Map) GetBool(key string) (bool, error)
```
GetBool is like Get except it always returns a boolean.

#### func (Map) GetObj

```go
func (scope Map) GetObj(key string) (AttributedObject, error)
```
GetObj is like Get except it always returns an AttributedObject.

#### func (Map) GetStr

```go
func (scope Map) GetStr(key string) (string, error)
```
GetStr is like Get except it always returns a string.

If the value is HTML, it is converted to a string.

#### func (Map) GetStrList

```go
func (scope Map) GetStrList(key string) ([]string, error)
```
GetStrList is like Get except it always returns a list of strings.

If the value is a `list{}` block, the list's values are returned, with
non-strings quietly filtered out.

If the value is a string, it is treated as a comma-separated list, and each item
is trimmed of prepending or suffixing whitespace.

#### func (*Map) Keys

```go
func (m *Map) Keys() []string
```
Keys returns a string of actual underlying map keys.

#### func (*Map) Map

```go
func (m *Map) Map() map[string]any
```
Map returns the actual underlying Go map.

#### func (*Map) OrderedKeys

```go
func (m *Map) OrderedKeys() []string
```
OrderedKeys returns a string of map keys in the order provided in the source.
Keys that were set internally (and not from quiki source code) are omitted.

#### func (Map) Set

```go
func (scope Map) Set(key string, value any) error
```
Set sets a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to write to a property of an object that does not support
properties, such as a string, Set returns an error.

#### func (Map) String

```go
func (b Map) String() string
```

#### func (Map) Unset

```go
func (scope Map) Unset(key string) error
```
Unset removes a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to unset a property of an object that does not support properties,
such as a string, Unset returns an error.

#### type ModelInfo

```go
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
```

ModelInfo represents metadata associated with a model.

#### type Page

```go
type Page struct {
	Source   string   // source content
	FilePath string   // Path to the .page file
	VarsOnly bool     // True if Parse() should only extract variables
	Opt      *PageOpt // page options

	Images    map[string][][]int   // references to images
	Models    map[string]ModelInfo // references to models
	PageLinks map[string][]int     // references to other pages

	Wiki     any  // only available during Parse() and HTML()
	Markdown bool // true if this is a markdown source

	Warnings []Warning // parser warnings
	Error    *Warning  // parser error, as an encodable Warning
}
```

Page represents a single page or article, generally associated with a .page
file. It provides the most basic public interface to parsing with the wikifier
engine.

#### func  NewPage

```go
func NewPage(filePath string) *Page
```
NewPage creates a page given its filepath.

#### func  NewPagePath

```go
func NewPagePath(filePath, name string) *Page
```
NewPagePath creates a page given its filepath and relative name.

#### func  NewPageSource

```go
func NewPageSource(source string) *Page
```
NewPageSource creates a page given some source code.

#### func (*Page) Author

```go
func (p *Page) Author() string
```
Author returns the page author's name, if any.

#### func (*Page) CSS

```go
func (p *Page) CSS() string
```
CSS generates and returns the CSS code for the page's inline styles.

#### func (*Page) CacheExists

```go
func (p *Page) CacheExists() bool
```
CacheExists is true if the page cache file exists.

#### func (*Page) CacheModified

```go
func (p *Page) CacheModified() time.Time
```
CacheModified returns the page cache file time.

#### func (*Page) CachePath

```go
func (p *Page) CachePath() string
```
CachePath returns the absolute path to the page cache file.

#### func (*Page) Categories

```go
func (p *Page) Categories() []string
```
Categories returns a list of categories the page belongs to.

#### func (*Page) Created

```go
func (p *Page) Created() time.Time
```
Created returns the page creation time.

#### func (*Page) Description

```go
func (p *Page) Description() string
```
Description returns the page description.

#### func (*Page) Draft

```go
func (p *Page) Draft() bool
```
Draft returns true if the page is marked as a draft.

#### func (*Page) Exists

```go
func (p *Page) Exists() bool
```
Exists is true if the page exists.

#### func (*Page) External

```go
func (p *Page) External() bool
```
External returns true if the page is outside the page directory as defined by
the configuration, with symlinks considered.

If `dir.wiki` isn't set, External is always true (since the page is not
associated with a wiki at all).

#### func (*Page) FmtTitle

```go
func (p *Page) FmtTitle() HTML
```
FmtTitle returns the page title, preserving any possible text formatting.

#### func (*Page) Generated

```go
func (p *Page) Generated() bool
```
Generated returns true if the page was auto-generated from some other source
content.

#### func (Page) Get

```go
func (scope Page) Get(key string) (any, error)
```
Get fetches a a value regardless of type.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to read a property of an object that does not support properties,
such as a string, Get returns an error.

If the key is valid but nothing exists at it, Get returns (nil, nil).

#### func (Page) GetBlock

```go
func (scope Page) GetBlock(key string) (block, error)
```
GetBlock is like Get except it always returns a block.

#### func (Page) GetBool

```go
func (scope Page) GetBool(key string) (bool, error)
```
GetBool is like Get except it always returns a boolean.

#### func (Page) GetObj

```go
func (scope Page) GetObj(key string) (AttributedObject, error)
```
GetObj is like Get except it always returns an AttributedObject.

#### func (Page) GetStr

```go
func (scope Page) GetStr(key string) (string, error)
```
GetStr is like Get except it always returns a string.

If the value is HTML, it is converted to a string.

#### func (Page) GetStrList

```go
func (scope Page) GetStrList(key string) ([]string, error)
```
GetStrList is like Get except it always returns a list of strings.

If the value is a `list{}` block, the list's values are returned, with
non-strings quietly filtered out.

If the value is a string, it is treated as a comma-separated list, and each item
is trimmed of prepending or suffixing whitespace.

#### func (*Page) HTML

```go
func (p *Page) HTML() HTML
```
HTML generates and returns the HTML code for the page. The page must be parsed
with Parse before attempting this method.

#### func (*Page) HTMLAndCSS

```go
func (p *Page) HTMLAndCSS() HTML
```
HTMLAndCSS generates and returns the HTML code for the page, including CSS.

#### func (*Page) Info

```go
func (p *Page) Info() PageInfo
```
Info returns the PageInfo for the page.

#### func (*Page) IsSymlink

```go
func (p *Page) IsSymlink() bool
```
IsSymlink returns true if the page is a symbolic link to another file within the
page directory. If it is symlinked to somewhere outside the page directory, it
is treated as a normal page rather than a redirect.

#### func (*Page) Keywords

```go
func (p *Page) Keywords() []string
```
Keywords returns the list of page keywords.

#### func (*Page) MarshalJSON

```go
func (p *Page) MarshalJSON() ([]byte, error)
```
MarshalJSON returns a JSON representation of the page. It includes the PageInfo,
HTML, and CSS.

#### func (*Page) Modified

```go
func (p *Page) Modified() time.Time
```
Modified returns the page modification time.

#### func (*Page) Name

```go
func (p *Page) Name() string
```
Name returns the resolved page name with extension.

This DOES take symbolic links into account. and DOES include the page prefix if
applicable. Any prefix will have forward slashes regardless of OS.

#### func (*Page) NameNE

```go
func (p *Page) NameNE() string
```
NameNE returns the resolved page name with No Extension.

#### func (*Page) OSName

```go
func (p *Page) OSName() string
```
OSName is like Name, except it uses the native path separator. It should be used
for file operations only.

#### func (*Page) OSNameNE

```go
func (p *Page) OSNameNE() string
```
OSNameNE is like NameNE, except it uses the native path separator. It should be
used for file operations only.

#### func (*Page) Parse

```go
func (p *Page) Parse() error
```
Parse opens the page file and attempts to parse it, returning any errors
encountered.

#### func (*Page) Path

```go
func (p *Page) Path() string
```
Path returns the absolute path to the page as resolved. If the path does not
resolve, returns an empty string.

#### func (*Page) Prefix

```go
func (p *Page) Prefix() string
```
Prefix returns the page prefix.

For example, for a page named a/b.page, this is a. For a page named a.page, this
is an empty string.

#### func (*Page) Preview

```go
func (p *Page) Preview() string
```
Preview returns a preview of the text on the page, up to 25 words or 150
characters. If the page has a Description, that is used instead of generating a
preview.

#### func (*Page) Redirect

```go
func (p *Page) Redirect() string
```
Redirect returns the location to which the page redirects, if any. This may be a
relative or absolute URL, suitable for use in a Location header.

#### func (*Page) RelName

```go
func (p *Page) RelName() string
```
RelName returns the unresolved page filename, with or without extension. This
does NOT take symbolic links into account. It is not guaranteed to exist.

#### func (*Page) RelNameNE

```go
func (p *Page) RelNameNE() string
```
RelNameNE returns the unresolved page name with No Extension, relative to the
page directory option. This does NOT take symbolic links into account. It is not
guaranteed to exist.

#### func (*Page) RelPath

```go
func (p *Page) RelPath() string
```
RelPath returns the unresolved file path to the page. It may be a relative or
absolute path. It is not guaranteed to exist.

#### func (*Page) SearchPath

```go
func (p *Page) SearchPath() string
```
SearchPath returns the absolute path to the page search text file.

#### func (Page) Set

```go
func (scope Page) Set(key string, value any) error
```
Set sets a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to write to a property of an object that does not support
properties, such as a string, Set returns an error.

#### func (*Page) Text

```go
func (p *Page) Text() string
```
Text generates and returns the rendered plain text for the page. The page must
be parsed with Parse before attempting this method.

#### func (*Page) Title

```go
func (p *Page) Title() string
```
Title returns the page title with HTML text formatting tags stripped.

#### func (*Page) TitleOrName

```go
func (p *Page) TitleOrName() string
```
TitleOrName returns the result of Title if available, otherwise that of Name.

#### func (Page) Unset

```go
func (scope Page) Unset(key string) error
```
Unset removes a value at the given key.

The key may be segmented to indicate properties of each object (e.g.
person.name).

If attempting to unset a property of an object that does not support properties,
such as a string, Unset returns an error.

#### func (*Page) Write

```go
func (p *Page) Write() error
```

#### type PageInfo

```go
type PageInfo struct {
	Path        string     `json:"-"`                   // absolute filepath
	File        string     `json:"file,omitempty"`      // name with extension, always with forward slashes
	FileNE      string     `json:"file_ne,omitempty"`   // name without extension, always with forward slashes
	Base        string     `json:"base,omitempty"`      // base name with extension
	BaseNE      string     `json:"base_ne,omitempty"`   // base name without extension
	Created     *time.Time `json:"created,omitempty"`   // creation time
	Modified    *time.Time `json:"modified,omitempty"`  // modify time
	Draft       bool       `json:"draft,omitempty"`     // true if page is marked as draft
	Generated   bool       `json:"generated,omitempty"` // true if page was generated from another source
	External    bool       `json:"external,omitempty"`  // true if page is outside the page directory
	Redirect    string     `json:"redirect,omitempty"`  // path page is to redirect to
	FmtTitle    HTML       `json:"fmt_title,omitempty"` // title with formatting tags
	Title       string     `json:"title,omitempty"`     // title without tags
	Author      string     `json:"author,omitempty"`    // author's name
	Description string     `json:"desc,omitempty"`      // description
	Keywords    []string   `json:"keywords,omitempty"`  // keywords
	Preview     string     `json:"preview,omitempty"`   // first 25 words or 150 chars. empty w/ description
	Warnings    []Warning  `json:"warnings,omitempty"`  // parser warnings
	Error       *Warning   `json:"error,omitempty"`     // parser error, as an encodable warning
}
```

PageInfo represents metadata associated with a page.

#### type PageOpt

```go
type PageOpt struct {
	Name         string // wiki name
	Logo         string // logo filename, relative to image dir
	MainPage     string // name of main page
	ErrorPage    string // name of error page
	Template     string // name of template
	MainRedirect bool   // redirect on main page rather than serve root
	Page         PageOptPage
	Host         PageOptHost
	Dir          PageOptDir
	Root         PageOptRoot
	Image        PageOptImage
	Category     PageOptCategory
	Search       PageOptSearch
	Link         PageOptLink
	External     map[string]PageOptExternal
	Navigation   []PageOptNavigation
}
```

PageOpt describes wiki/website options to a Page.

#### type PageOptCategory

```go
type PageOptCategory struct {
	PerPage int
}
```

PageOptCategory describes wiki category options.

#### type PageOptCode

```go
type PageOptCode struct {
	Lang  string
	Style string
}
```

PageOptCode describes options for `code{}` blocks.

#### type PageOptDir

```go
type PageOptDir struct {
	Wiki     string // path to wiki root directory
	Image    string // Deprecated: path to image directory
	Category string // Deprecated: path to category directory
	Page     string // Deprecated: path to page directory
	Model    string // Deprecated: path to model directory
	Markdown string // Deprecated: path to markdown directory
	Cache    string // Deprecated: path to cache directory
}
```

PageOptDir describes actual filepaths to wiki resources.

#### type PageOptExternal

```go
type PageOptExternal struct {
	Name string              // long name (e.g. Wikipedia)
	Root string              // wiki page root (no trailing slash)
	Type PageOptExternalType // wiki type
}
```

PageOptExternal describes an external wiki that we can use for link targets.

#### type PageOptExternalType

```go
type PageOptExternalType string
```

PageOptExternalType describes

#### type PageOptHost

```go
type PageOptHost struct {
	Wiki string // HTTP host for the wiki
}
```

PageOptHost describes HTTP hosts for a wiki.

#### type PageOptImage

```go
type PageOptImage struct {
	Retina     []int
	SizeMethod string
	Calc       func(file string, width, height int, page *Page) (w, h int, fullSize bool)
	Sizer      func(file string, width, height int, page *Page) (path string)
}
```

PageOptImage describes wiki imaging options.

#### type PageOptLink

```go
type PageOptLink struct {
	ParseInternal PageOptLinkFunction // internal page links
	ParseExternal PageOptLinkFunction // external wiki page links
	ParseCategory PageOptLinkFunction // category links
}
```

PageOptLink describes functions to assist with link targets.

#### type PageOptLinkFunction

```go
type PageOptLinkFunction func(page *Page, opts *PageOptLinkOpts)
```

A PageOptLinkFunction sanitizes a link target.

#### type PageOptLinkOpts

```go
type PageOptLinkOpts struct {
	Ok             *bool   // func sets to true if the link is valid
	Target         *string // func sets to overwrite the link target
	Tooltip        *string // func sets tooltip to display
	DisplayDefault *string // func sets default text to display (if no pipe)
	*FmtOpt                // formatter options available to func
}
```

PageOptLinkOpts contains options passed to a PageOptLinkFunction.

#### type PageOptNavigation

```go
type PageOptNavigation struct {
	Link    string // link
	Display string // text to display
}
```

PageOptNavigation represents an ordered navigation item.

#### type PageOptPage

```go
type PageOptPage struct {
	EnableTitle bool        // enable page title headings
	EnableCache bool        // enable page caching
	ForceGen    bool        // force generation of page even if unchanged
	Code        PageOptCode // `code{}` block options
}
```

PageOptPage describes option relating to a page.

#### type PageOptRoot

```go
type PageOptRoot struct {
	Wiki     string // wiki root path
	Image    string // image root path
	Category string // category root path
	Page     string // page root path
	File     string // file index path
	Ext      string // full external wiki prefix
}
```

PageOptRoot describes HTTP paths to wiki resources.

#### type PageOptSearch

```go
type PageOptSearch struct {
	Enable bool
}
```

PageOptSearch describes wiki search options.

#### type ParserError

```go
type ParserError struct {
	Pos Position
	Err error
}
```

ParserError represents an error in parsing with positional info.

#### func (*ParserError) Error

```go
func (e *ParserError) Error() string
```

#### func (*ParserError) Unwrap

```go
func (e *ParserError) Unwrap() error
```

#### type Position

```go
type Position struct {
	Line, Column int
}
```

Position represents a line and column position within a quiki source file.

#### func (*Position) MarshalJSON

```go
func (pos *Position) MarshalJSON() ([]byte, error)
```
MarshalJSON encodes the position to `[line, column]`.

#### func (Position) String

```go
func (pos Position) String() string
```
String returns the position as `{line column}`.

#### func (*Position) UnmarshalJSON

```go
func (pos *Position) UnmarshalJSON(data []byte) error
```
UnmarshalJSON decodes the position from `[line, column]`.

#### type Warning

```go
type Warning struct {
	Message string   `json:"message"`
	Pos     Position `json:"position"`
}
```

Warning represents a warning on a page.

#### func (Warning) Log

```go
func (w Warning) Log(path string)
```
