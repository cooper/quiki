package wikifier

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var nonAlphaRegex = regexp.MustCompile(`[^\w\.\-\/]`)

// Represents a quiki value type.
type valueType int

// These are the value types in a quiki page model.
const (
	valueTypeString  valueType = iota // String
	valueTypeBlock                    // Block
	valueTypeHTML                     // Raw HTML
	valueTypeElement                  // HTML element
	valueTypeMixed                    // []any with a mixture of these
)

// Returns a quiki valueType, given a value. Accepted types are string,
// block, HTML, element, or an []any with mixed types.
// If i is none of these, returns -1.
func getValueType(i any) valueType {
	switch i.(type) {
	case HTML:
		return valueTypeHTML
	case string:
		return valueTypeString
	case block:
		return valueTypeBlock
	case element:
		return valueTypeElement
	case []any:
		return valueTypeMixed
	default:
		return -1
	}
}

// Prepares a quiki value for representation as HTML.
//
// Strings are formatted with quiki's text formatter.
// Blocks are converted to elements.
// Preformatted HTML is left as-is.
func prepareForHTML(value any, b block, pos Position) any {
	switch v := value.(type) {
	case string:
		value = format(b, v, pos)
	case block:
		// copied because it might be iterated over in for{} blocks
		copied := v.el().copy()
		v.html(b.page(), copied)
		value = copied
	case []any:
		newValues := make([]any, len(v))
		for idx, val := range v {
			newValues[idx] = prepareForHTML(val, b, pos)
		}
		value = newValues
	case HTML:
	default:
		panic("not sure what to do with this")
	}
	return value
}

// convert wikifier values to human-readable form
func humanReadableValue(i any) string {
	switch v := i.(type) {

	// FIXME: recursions possible?

	// list
	case []any:
		hrValues := make([]string, len(v))
		for i, val := range v {
			hrValues[i] = humanReadableValue(val)
		}
		return strings.Join(hrValues, " ")

	// nothing
	case nil:
		return ""

	// string
	case string:
		return `'` + strings.TrimSpace(strings.ReplaceAll(v, "\n", "\u2424")) + `'`

	// boolean
	case bool:
		if v {
			return "true"
		}
		return "false"

	// block
	case block:
		return v.String()

	// element
	case element:
		return "<" + v.tag() + ">..."

	// something else
	default:
		return fmt.Sprintf("%v", v)
	}
}

// fix a value before storing it in a list or map
// this returns either a string, block, or []any of both
// strings next to each other are merged; empty strings are removed
func fixValuesForStorage(values []any, blockMaybe block, pos Position, fmtText bool) any {

	// no items
	if len(values) == 0 {
		return nil
	}

	// one value in; one value out!
	if len(values) == 1 {
		return fixSingleValue(values[0], blockMaybe, pos, fmtText)
	}

	// multiple values
	var valuesToStore []any
	var lastValue any
	for _, value := range values {

		// fix this value; then skip it if it's nothin
		value = fixSingleValue(value, blockMaybe, pos, fmtText)
		if value == nil {
			continue
		}

		// if this is a string/HTML and the previous one was too, combine them
		thisStr, isStr := value.(string)
		lastStr, lastWasStr := lastValue.(string)
		thisHTML, isHTML := value.(HTML)
		lastHTML, lastWasHTML := lastValue.(HTML)
		if isHTML && lastWasHTML {
			valuesToStore[len(valuesToStore)-1] = HTML(lastHTML + thisHTML)
		} else if isStr && lastWasStr {
			valuesToStore[len(valuesToStore)-1] = lastStr + thisStr
		} else {
			valuesToStore = append(valuesToStore, value)
		}

		lastValue = value
	}

	// ended up with nothing
	if len(valuesToStore) == 0 {
		return nil
	}

	// ended up with just one
	if len(valuesToStore) == 1 {
		return valuesToStore[0]
	}

	// mix of strings and blocks
	return valuesToStore
}

func fixSingleValue(value any, blockMaybe block, pos Position, fmtText bool) any {
	switch v := value.(type) {
	case HTML:
		return v
	case string:
		v = strings.Trim(v, "\t ") // remove non-newline spaces
		if v == "" {
			return nil
		}
		if fmtText && blockMaybe != nil {
			return format(blockMaybe, v, pos)
		}
		return v
	case block:
		// just in case

		if v == nil {
			return nil
		}
		return value
	default:
		panic("somehow a non-string and non-block value got into a map or list")
	}
}

// UniqueFilesInDir recursively scans a directory for files matching the
// requested extensions, resolves symlinks, and returns a list of
// unique files. That is, if more than one link resolves to the same thing
// (as is the case for quiki page redirects), there is only one occurrence
// in the output.
func UniqueFilesInDir(dir string, extensions []string, thisDirOnly bool) ([]string, error) {
	uniqueFiles := make(map[string]string)

	// nothin in, nothin out
	if dir == "" {
		return nil, nil
	}

	// no need for trailing /
	if dir[len(dir)-1] == filepath.Separator {
		dir = dir[:len(dir)-1]
	}

	dirAbs, _ := filepath.Abs(dir)

	var doDir func(pfx string)
	doDir = func(pfx string) {
		dir := filepath.Join(dir, pfx)

		// can't open directory
		files, err := os.ReadDir(dir)
		if err != nil {
			// TODO: report the error somehow
			return
		}

		// check each file
		for _, file := range files {
			path := filepath.Join(dir, file.Name())

			// skip hidden files
			if file.Name()[0] == '.' {
				continue
			}

			// this is a symlink; follow it if it's a directory
			isDir := file.IsDir()
			if !isDir && file.Type()&os.ModeSymlink != 0 {
				stat, err := os.Stat(path)
				if err == nil && stat.IsDir() {
					isDir = true
				}
			}

			// this is a directory
			if isDir {
				if !thisDirOnly {
					doDir(pfx + file.Name() + string(filepath.Separator))
				}
				continue
			}

			// find extension
			ext := ""
			if lastDot := strings.LastIndexByte(file.Name(), '.'); lastDot != -1 && lastDot < len(file.Name())-1 {
				ext = file.Name()[lastDot+1:]
			}

			// skip files without desired extension
			skip := true
			for _, acceptable := range extensions {
				if ext == acceptable {
					skip = false
					break
				}
			}
			if skip {
				continue
			}

			// resolve symlinks
			symlinkOrOrig, err := filepath.EvalSymlinks(path)
			if err != nil {
				continue
			}

			// resolve absolute path
			abs, err := filepath.Abs(symlinkOrOrig)
			if err != nil {
				// TODO: report the error
				continue
			}

			// use the basename of the resolved path only if the target
			// file is in the same directory; otherwise use the original path
			filename := path
			if rel, err := filepath.Rel(dirAbs, abs); err == nil {
				if strings.IndexByte(rel, filepath.Separator) == -1 {
					filename = abs
				}
			}
			filename = filepath.Base(filename)

			// remember this file
			uniqueFiles[strings.ToLower(pfx+filename)] = pfx + filename
		}
	}

	doDir("")

	// convert back to list
	i := 0
	unique := make([]string, len(uniqueFiles))
	for _, name := range uniqueFiles {
		unique[i] = name
		i++
	}

	return unique, nil
}

// PageName returns a clean page name.
func PageName(name string) string {
	return PageNameExt(name, "")
}

// PageNameNE returns a clean page name with No Extension.
func PageNameNE(name string) string {
	// TODO: make this less ugly
	name = strings.TrimSuffix(PageName(name), ".page")
	name = strings.TrimSuffix(name, ".model")
	name = strings.TrimSuffix(name, ".conf")
	name = strings.TrimSuffix(name, ".md")
	return name
}

// PageNameExt returns a clean page name with the provided extension.
func PageNameExt(name, ext string) string {
	// 'Some Article' -> 'some_article.page'

	if ext == "" {
		ext = ".page"
	}

	// convert non-alphanumerics to _
	name = PageNameLink(name)

	// append the extension if it isn't already there
	lastDot := strings.LastIndexByte(name, '.')
	if lastDot != -1 && lastDot < len(name)-1 {
		existing := name[lastDot:]
		if existing != ".page" && existing != ".model" && existing != ".conf" && existing != ".md" {
			// TODO: make above prettier
			name += ext
		}
	} else {
		name += ext
	}

	return name
}

// PageNameLink returns a clean page name without the extension.
func PageNameLink(name string) string {
	name = strings.TrimSpace(name)
	// 'Some Article' -> 'Some_Article'

	// don't waste any time
	if name == "" {
		return name
	}

	// just in case, convert any native path separators to /
	name = filepath.ToSlash(name)

	// replace non-alphanumerics with _
	name = nonAlphaRegex.ReplaceAllString(name, "_")

	return name
}

// CategoryName returns a clean category name.
func CategoryName(name string) string {
	name = PageNameLink(name)
	if !strings.HasSuffix(name, ".cat") {
		return name + ".cat"
	}
	return name
}

// CategoryNameNE returns a clean category with No Extension.
func CategoryNameNE(name string) string {
	return strings.TrimSuffix(PageNameLink(name), ".cat")
}

// ModelName returns a clean model name.
func ModelName(name string) string {
	return PageNameExt(name, ".model")
}

// MakeDir creates directories recursively.
func MakeDir(dir, name string) {
	pfx := filepath.Dir(name)
	os.MkdirAll(filepath.Join(dir, pfx), 0755)
}

// ScaleString returns a string of scaled image names for use in srcset.
func ScaleString(name string, retina []int) string {

	// find image name and extension
	imageName, ext := name, ""
	if lastDot := strings.LastIndexByte(name, '.'); lastDot != -1 {
		imageName = name[:lastDot]
		ext = name[lastDot:]
	}

	// rewrite a.jpg to a@2x.jpg
	scales := make([]string, len(retina))
	for i, scale := range retina {
		scaleStr := strconv.Itoa(scale) + "x"
		scales[i] = imageName + "@" + scaleStr + ext + " " + scaleStr
	}

	return strings.Join(scales, ", ")
}
