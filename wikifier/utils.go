package wikifier

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Represents a quiki value type.
type valueType int

// These are the value types in a quiki page model.
const (
	valueTypeString  valueType = 0 // String
	valueTypeBlock                 // Block
	valueTypeHTML                  // Raw HTML
	valueTypeElement               // HTML element
	valueTypeMixed                 // []interface{} with a mixture of these
)

// Returns a quiki valueType, given a value. Accepted types are string,
// block, HTML, element, or an []interface{} with mixed types.
// If i is none of these, returns -1.
func getValueType(i interface{}) valueType {
	switch i.(type) {
	case string:
		return valueTypeString
	case block:
		return valueTypeBlock
	case HTML:
		return valueTypeHTML
	case element:
		return valueTypeElement
	case []interface{}:
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
//
func prepareForHTML(value interface{}, page *Page, pos position) interface{} {
	switch v := value.(type) {
	case string:
		value = page.parseFormattedTextOpts(v, &formatterOptions{pos: pos})
	case block:
		v.html(page, v.el())
		value = v.el()
	case []interface{}:
		newValues := make([]interface{}, len(v))
		for idx, val := range v {
			newValues[idx] = prepareForHTML(val, page, pos)
		}
		value = newValues
	case HTML:
	default:
		panic("not sure what to do with this")
	}
	return value
}

func pageNameLink(s string) string {
	return ""
}

// convert wikifier values to human-readable form
func humanReadableValue(i interface{}) string {
	switch v := i.(type) {

	// FIXME: recursions possible?

	// list
	case []interface{}:
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
		return `'` + strings.TrimSpace(v) + `'`

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
// this returns either a string, block, or []interface{} of both
// strings next to each other are merged; empty strings are removed
func fixValuesForStorage(values []interface{}) interface{} {

	// no items
	if len(values) == 0 {
		return nil
	}

	// one value in; one value out!
	if len(values) == 1 {
		return fixSingleValue(values[0])
	}

	// multiple values
	var valuesToStore []interface{}
	var lastValue interface{}
	for _, value := range values {

		// fix this value; then skip it if it's nothin
		value = fixSingleValue(value)
		if value == nil {
			continue
		}

		// if this is a string and the previous one was too, combine them
		thisStr, isStr := value.(string)
		lastStr, lastWasStr := lastValue.(string)
		if isStr && lastWasStr {
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

func fixSingleValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
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

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func UniqueFilesInDir(dir string, extensions []string, thisDirOnly bool) ([]string, error) {
	uniqueFiles := make(map[string]string)

	// nothin in, nothin out
	if dir == "" {
		return nil, nil
	}

	// no need for trailing /
	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}

	dirAbs, _ := filepath.Abs(dir)

	var doDir func(pfx string)
	doDir = func(pfx string) {
		dir := dir + "/" + pfx
		fmt.Println("doDir", pfx, dir)

		// can't open directory
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			// TODO: report the error somehow
			fmt.Println(err)
			return
		}

		// check each file
		for _, file := range files {
			path := dir + file.Name()

			// skip hidden files
			if file.Name()[0] == '.' {
				fmt.Println("hidden", file.Name())
				continue
			}

			// this is a directory
			if file.IsDir() {
				if !thisDirOnly {
					doDir(pfx + file.Name() + "/")
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
				fmt.Println("check", file.Name(), ext, "==", acceptable)

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
			fmt.Println("sym", path, symlinkOrOrig, err)
			if err != nil {
				continue
			}

			// resolve absolute path
			abs, err := filepath.Abs(symlinkOrOrig)
			fmt.Println("abs", path, abs, err)
			if err != nil {
				// TODO: report the error
				continue
			}

			// use the basename of the resolved path only if the target
			// file is in the same directory; otherwise use the original path
			filename := path
			a, b := filepath.Rel(dirAbs, abs)
			fmt.Println("rel", path, abs, a, b)

			if rel, err := filepath.Rel(dirAbs, abs); err == nil {
				if strings.IndexByte(rel, '/') == -1 {
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
