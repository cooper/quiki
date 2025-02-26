package webserver

import (
	"net/http"
)

type ServeMux struct {
	*http.ServeMux
	routes []Route
}

type Route struct {
	Pattern     string
	Description string
}

func NewServeMux() *ServeMux {
	return &ServeMux{http.NewServeMux(), nil}
}

// Register registers the handler for the given pattern and adds to routes.
func (m *ServeMux) Register(pattern, description string, handler http.Handler) {
	m.routes = append(m.routes, Route{pattern, description})
	m.ServeMux.Handle(pattern, handler)
}

// RegisterFunc registers the handler function for the given pattern and adds to routes.
func (m *ServeMux) RegisterFunc(pattern, description string, handler func(http.ResponseWriter, *http.Request)) {
	m.routes = append(m.routes, Route{pattern, description})
	m.ServeMux.HandleFunc(pattern, handler)
}

// GetRoutes returns the registered routes.
func (m *ServeMux) GetRoutes() []Route {
	return m.routes
}
