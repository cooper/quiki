package wikifier

import "errors"

type attributedObject interface {
	SetStr(key, value string) error
	SetObj(key string, value attributedObject) error
	GetErr(key string) (interface{}, error)
	GetStr(key string) (string, error)
	GetObj(key string) (attributedObject, error)
}

type variableScope struct {
	vars map[string]interface{}
}

func newVariableScope() *variableScope {
	return &variableScope{make(map[string]interface{})}
}

func (scope *variableScope) SetStr(key, value string) error {
	scope.vars[key] = value
	return nil
}

func (scope *variableScope) SetObj(key string, value attributedObject) error {
	scope.vars[key] = value
	return nil
}

func (scope *variableScope) GetErr(key string) (interface{}, error) {
	if val, exist := scope.vars[key]; exist {
		return val, nil
	}
	return nil, errors.New("nonexistent key")
}

func (scope *variableScope) GetStr(key string) (string, error) {
	obj, err := scope.GetErr(key)
	if err != nil {
		return "", err
	}
	if str, ok := obj.(string); ok {
		return str, nil
	}
	return "", errors.New("not a string")
}

func (scope *variableScope) GetObj(key string) (attributedObject, error) {
	obj, err := scope.GetErr(key)
	if err != nil {
		return nil, err
	}
	if obj, ok := obj.(attributedObject); ok {
		return obj, nil
	}
	return nil, errors.New("not an object")
}
