package wikifier

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// An AttributedObject is any object on which you can set and retrieve attributes.
//
// For example, a Page is an attributed object since it contains variables.
// Likewise, a Map is an attributed object because it has named properties.
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

	// internal use
	mainBlock() block
	setOwn(key string, value any)
	getOwn(key string) any
	get(key string) any
	unsetOwn(key string)
}

type variableScope struct {
	vars   map[string]any
	parent *variableScope
}

// newVariableScope creates a variable scope
func newVariableScope() *variableScope {
	return &variableScope{vars: make(map[string]any)}
}

// newVariableScopeWithParent creates a variable scope with a parent scope
func newVariableScopeWithParent(parent *variableScope) *variableScope {
	return &variableScope{vars: make(map[string]any), parent: parent}
}

func (scope *variableScope) mainBlock() block {
	return nil
}

// Set sets a value at the given key.
//
// The key may be segmented to indicate properties of each object
// (e.g. person.name).
//
// If attempting to write to a property of an object that does not
// support properties, such as a string, Set returns an error.
func (scope *variableScope) Set(key string, value any) error {
	var where AttributedObject = scope

	// split into parts
	parts := strings.Split(key, ".")
	setting, parts := parts[len(parts)-1], parts[:len(parts)-1]

	for _, name := range parts {

		// fetch the next object
		newWhere, err := where.GetObj(name)
		if err != nil {
			return err
		}

		// this location doesn't exist; make a new map
		if newWhere == nil {
			newWhere = NewMap(where.mainBlock())
			where.setOwn(name, newWhere)
			// TODO: maybe somehow include some positioning info here?
		}

		where = newWhere
	}

	// finally, set it on the last object
	where.setOwn(setting, value)
	return nil
}

// Unset removes a value at the given key.
//
// The key may be segmented to indicate properties of each object
// (e.g. person.name).
//
// If attempting to unset a property of an object that does not
// support properties, such as a string, Unset returns an error.
func (scope *variableScope) Unset(key string) error {
	var where AttributedObject = scope

	// split into parts
	parts := strings.Split(key, ".")
	setting, parts := parts[len(parts)-1], parts[:len(parts)-1]

	for _, name := range parts {
		newWhere, err := where.GetObj(name)
		if err != nil {
			return err
		}

		// no error, but there's nothing there
		if newWhere == nil {
			return nil
		}

		where = newWhere
	}

	// finally, unset it on the last object
	where.unsetOwn(setting)
	return nil
}

// Get fetches a a value regardless of type.
//
// The key may be segmented to indicate properties of each object
// (e.g. person.name).
//
// If attempting to read a property of an object that does not
// support properties, such as a string, Get returns an error.
//
// If the key is valid but nothing exists at it, Get returns (nil, nil).
func (scope *variableScope) Get(key string) (any, error) {
	var where AttributedObject = scope

	parts := strings.Split(key, ".")
	setting, parts := parts[len(parts)-1], parts[:len(parts)-1]

	for _, name := range parts {
		newWhere, err := where.GetObj(name)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}

		// no error, but there's nothing there
		if newWhere == nil {
			return nil, nil //fmt.Errorf("Get(%s): '%s' does not exist", key, parts[i])
		}

		where = newWhere
	}

	return where.get(setting), nil
}

// GetStr is like Get except it always returns a string.
//
// If the value is HTML, it is converted to a string.
func (scope *variableScope) GetStr(key string) (string, error) {
	val, err := scope.Get(key)
	if err != nil {
		return "", err
	}

	// there is nothing here
	if val == nil {
		return "", nil
	}

	// something is here, so it best be a string
	if str, ok := val.(string); ok {
		return str, nil
	} else if html, ok := val.(HTML); ok {
		return string(html), nil
	}

	// not what we asked for
	return "", errors.New("not a string (" + humanReadableValue(val) + ")")
}

// GetBool is like Get except it always returns a boolean.
func (scope *variableScope) GetBool(key string) (bool, error) {
	val, err := scope.Get(key)
	if err != nil {
		return false, err
	}

	// there is nothing here
	if val == nil {
		return false, nil
	}

	// something is here, so it best be a bool
	if b, ok := val.(bool); ok {
		return b, nil
	}

	// not what we asked for
	return false, errors.New("not a boolean")
}

// GetObj is like Get except it always returns an AttributedObject.
func (scope *variableScope) GetObj(key string) (AttributedObject, error) {
	obj, err := scope.Get(key)
	if err != nil {
		return nil, err
	}

	// there is nothing here
	if obj == nil {
		return nil, nil
	}

	// something is here, so it best be an AttributedObject
	if aObj, ok := obj.(AttributedObject); ok {
		return aObj, nil
	}

	// not what we asked for
	return nil, errors.New("not an object")
}

// GetBlock is like Get except it always returns a block.
func (scope *variableScope) GetBlock(key string) (block, error) {
	obj, err := scope.Get(key)
	if err != nil {
		return nil, err
	}

	// there is nothing here
	if obj == nil {
		return nil, nil
	}

	// something is here, so it best be an AttributedObject
	if blk, ok := obj.(block); ok {
		return blk, nil
	}

	// not what we asked for
	return nil, errors.New("not a block")
}

// GetStrList is like Get except it always returns a list of strings.
//
// If the value is a `list{}` block, the list's values are returned,
// with non-strings quietly filtered out.
//
// If the value is a string, it is treated as a comma-separated list,
// and each item is trimmed of prepending or suffixing whitespace.
func (scope *variableScope) GetStrList(key string) ([]string, error) {
	val, err := scope.Get(key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {

	// list{} block
	case *List:
		var list []string
		for _, entry := range v.list {
			if entry.typ == valueTypeString {
				list = append(list, entry.value.(string))
			} else if entry.typ == valueTypeHTML {
				list = append(list, string(entry.value.(HTML)))
			}
		}
		return list, nil

	// comma-separated list
	case string, HTML:

		// get string
		var str string
		if html, ok := v.(HTML); ok {
			str = string(html)
		} else {
			str = v.(string)
		}

		var list []string
		for _, item := range strings.Split(str, ",") {

			// trim whitespace
			item = strings.TrimSpace(item)

			// nothing left/blank entry
			if item == "" {
				continue
			}

			// add it
			list = append(list, item)
		}
		return list, nil
	}

	// something else
	return nil, errors.New("not a list{} or comma-separated list")
}

// GetInt is like Get except it always returns an integer.
// If the value is a string, it attempts to parse it as an integer.
// Returns (0, false, nil) if the key doesn't exist.
// Returns (value, true, nil) if the key exists and can be parsed.
// Returns (0, false, error) if the key exists but cannot be parsed.
func (scope *variableScope) GetInt(key string) (int, bool, error) {
	val, err := scope.Get(key)
	if err != nil {
		return 0, false, err
	}

	// there is nothing here
	if val == nil {
		return 0, false, nil
	}

	// if it's already an int, return it
	if i, ok := val.(int); ok {
		return i, true, nil
	}

	// if it's a string, try to parse it
	if str, ok := val.(string); ok {
		if str == "" {
			return 0, true, nil
		}
		parsed, err := strconv.Atoi(str)
		if err != nil {
			return 0, false, errors.Wrap(err, "cannot parse as integer")
		}
		return parsed, true, nil
	}

	// not what we asked for
	return 0, false, errors.New("not an integer or parseable string")
}

// GetDuration is like Get except it always returns a time.Duration.
// If the value is a string, it attempts to parse it as a duration (e.g., "30s", "5m", "1h").
// Returns (0, false, nil) if the key doesn't exist.
// Returns (value, true, nil) if the key exists and can be parsed.
// Returns (0, false, error) if the key exists but cannot be parsed.
func (scope *variableScope) GetDuration(key string) (time.Duration, bool, error) {
	val, err := scope.Get(key)
	if err != nil {
		return 0, false, err
	}

	// there is nothing here
	if val == nil {
		return 0, false, nil
	}

	// if it's already a duration, return it
	if d, ok := val.(time.Duration); ok {
		return d, true, nil
	}

	// if it's a string, try to parse it
	if str, ok := val.(string); ok {
		if str == "" {
			return 0, true, nil
		}
		parsed, err := time.ParseDuration(str)
		if err != nil {
			return 0, false, errors.Wrap(err, "cannot parse as duration (use format like '30s', '5m', '1h')")
		}
		return parsed, true, nil
	}

	// not what we asked for
	return 0, false, errors.New("not a duration or parseable string")
}

// INTERNAL

// set own property
func (scope *variableScope) setOwn(key string, value any) {
	scope.vars[key] = value
}

// unset own property
// note this does not unset properties of parent scopes
func (scope *variableScope) unsetOwn(key string) {
	delete(scope.vars, key)
}

// fetch own property
func (scope *variableScope) getOwn(key string) any {
	if val, exist := scope.vars[key]; exist {
		return val
	}
	return nil
}

// fetch own or parent scope property
func (scope *variableScope) get(key string) any {
	if val, exist := scope.vars[key]; exist {
		return val
	}
	if scope.parent != nil {
		return scope.parent.get(key)
	}
	return nil
}
