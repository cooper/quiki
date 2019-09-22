package wikifier

type variableScope struct {
	vars map[string]interface{}
}

func newVariableScope() *variableScope {
	return &variableScope{make(map[string]interface{})}
}

func (scope *variableScope) Set(key string, value interface{}) error {

	return nil
}

func (scope *variableScope) GetErr(key string, value interface{}) (interface{}, error) {
	return scope.vars[key], nil
}
