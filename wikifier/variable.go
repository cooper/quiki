package wikifier

import (
	"errors"
	"fmt"
	"strings"
)

type attributedObject interface {
	Get(key string) (interface{}, error)
	GetBool(key string) (bool, error)
	GetStr(key string) (string, error)
	GetObj(key string) (attributedObject, error)

	Set(key string, value interface{}) (interface{}, error)
	SetBool(key string, value bool) error
	SetStr(key, value string) error
	SetObj(key string, value attributedObject) error

	// internal use
	mainBlock() block
	setOwn(key string, value interface{})
	getOwn(key string) interface{}
}

type variableScope struct {
	vars map[string]interface{}
}

func newVariableScope() *variableScope {
	return &variableScope{make(map[string]interface{})}
}

func (scope *variableScope) mainBlock() block {
	return nil
}

func (scope *variableScope) Set(key string, value interface{}) (interface{}, error) {
	var where attributedObject = scope
	// my @parts   = split /\./, $var;
	// my $setting = pop @parts;

	parts := strings.Split(key, ".")
	setting, parts := parts[len(parts)-1], parts[:len(parts)-1]

	// while (length($var = shift @parts)) {
	for _, name := range parts {

		//     my ($new_where, $err) = _get_attr($where, $var);
		//     return (undef, $err) if $err
		newWhere, err := where.GetObj(name)
		if err != nil {
			return "", err
		}

		// this location doesn't exist; make a new map
		if newWhere == nil {
			newWhere = NewMap(where.mainBlock())
			// TODO: maybe somehow include some positioning info here?
		}

		where = newWhere
	}

	where.setOwn(setting, value)
	return value, nil
}

func (scope *variableScope) SetBool(key string, value bool) error {
	scope.vars[key] = value
	return nil
}

func (scope *variableScope) SetStr(key, value string) error {
	scope.vars[key] = value
	return nil
}

func (scope *variableScope) SetObj(key string, value attributedObject) error {
	scope.vars[key] = value
	return nil
}

// fetch a variable regardless of type
// only fails if attempting to fetch attributes on non attributed value
// does not fail due to the absence of a value
func (scope *variableScope) Get(key string) (interface{}, error) {
	var where attributedObject = scope

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

// fetch the string value of a variable
// fails only if a non-string value is present
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

// fetch the string value of a variable
// fails only if a non-bool value is present
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

// fetch the object value of a variable
// fails only if a non-object value is present
func (scope *variableScope) GetObj(key string) (attributedObject, error) {
	obj, err := scope.Get(key)
	if err != nil {
		return nil, err
	}

	// there is nothing here
	if obj == nil {
		return nil, nil
	}

	// something is here, so it best be an attributedObject
	if aObj, ok := obj.(attributedObject); ok {
		return aObj, nil
	}

	// not what we asked for
	return nil, errors.New("not an object")
}

// INTERNAL

func (scope *variableScope) setOwn(key string, value interface{}) {
	scope.vars[key] = value
}

func (scope *variableScope) getOwn(key string) interface{} {
	if val, exist := scope.vars[key]; exist {
		return val
	}
	return nil
}
