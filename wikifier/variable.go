package wikifier

import (
	"errors"
	"fmt"
	"strings"
)

// An AttributedObject is any object on which you can set and retrieve attributes.
//
// For example, a Page is an attributed object since it contains variables.
// Likewise, a Map is an attributed object because it has named properties.
//
type AttributedObject interface {

	// getters
	Get(key string) (interface{}, error)
	GetBool(key string) (bool, error)
	GetStr(key string) (string, error)
	GetObj(key string) (AttributedObject, error)

	// setters
	Set(key string, value interface{}) error

	// internal use
	mainBlock() block
	setOwn(key string, value interface{})
	getOwn(key string) interface{}
}

type variableScope struct {
	vars map[string]interface{}
}

// newVariableScope creates a variable scope
func newVariableScope() *variableScope {
	return &variableScope{make(map[string]interface{})}
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
//
func (scope *variableScope) Set(key string, value interface{}) error {
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
			// TODO: maybe somehow include some positioning info here?
		}

		where = newWhere
	}

	// finally, set it on the last object
	where.setOwn(setting, value)
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
//
func (scope *variableScope) Get(key string) (interface{}, error) {
	var where AttributedObject = scope

	parts := strings.Split(key, ".")
	setting, parts := parts[len(parts)-1], parts[:len(parts)-1]

	for i, name := range parts {
		newWhere, err := where.GetObj(name)
		if err != nil {
			return nil, err
		}

		// no error, but there's nothing there
		if newWhere == nil {
			return nil, fmt.Errorf("Get(%s): '%s' does not exist", key, parts[i])
		}

		where = newWhere
	}

	return where.getOwn(setting), nil
}

// GetStr is like Get except it always returns a string.
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
	}

	// not what we asked for
	return "", errors.New("not a string")
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

// INTERNAL

// set own property
func (scope *variableScope) setOwn(key string, value interface{}) {
	scope.vars[key] = value
}

// fetch own property
func (scope *variableScope) getOwn(key string) interface{} {
	if val, exist := scope.vars[key]; exist {
		return val
	}
	return nil
}
