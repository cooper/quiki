// Package router provides a lightweight, dynamic HTTP router for quiki.
package router

import (
	"net/http"
	"strings"
	"sync"
)

// Router is a lightweight HTTP router that supports dynamic route addition/removal.
type Router struct {
	mu     sync.RWMutex
	routes map[string]*Route
	static map[string]http.Handler // for exact match routes like "/admin"
}

// Route represents a registered route with its handler and metadata.
type Route struct {
	Pattern     string
	Description string
	Handler     http.Handler
	WikiName    string // for wiki-specific routes
}

// Params contains path parameters extracted from the URL.
type Params struct {
	values map[string]string
}

// Get returns the value of a path parameter.
func (p *Params) Get(key string) string {
	if p.values == nil {
		return ""
	}
	return p.values[key]
}

// New creates a new router instance.
func New() *Router {
	return &Router{
		routes: make(map[string]*Route),
		static: make(map[string]http.Handler),
	}
}

// Handle registers a handler for the given pattern.
func (r *Router) Handle(pattern, description string, handler http.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := &Route{
		Pattern:     pattern,
		Description: description,
		Handler:     handler,
	}

	// if pattern has no wildcards AND doesn't end with /, treat as static route for fast lookup
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, ":") && !strings.HasSuffix(pattern, "/") {
		r.static[pattern] = handler
	} else {
		r.routes[pattern] = route
	}
}

// HandleFunc registers a handler function for the given pattern.
func (r *Router) HandleFunc(pattern, description string, handler func(http.ResponseWriter, *http.Request)) {
	r.Handle(pattern, description, http.HandlerFunc(handler))
}

// HandleWiki registers handlers for a specific wiki with automatic cleanup support.
func (r *Router) HandleWiki(wikiName, pattern, description string, handler http.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := &Route{
		Pattern:     pattern,
		Description: description,
		Handler:     handler,
		WikiName:    wikiName,
	}

	r.routes[pattern] = route
}

// RemoveWiki removes all routes associated with a specific wiki.
func (r *Router) RemoveWiki(wikiName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// remove from routes map
	for pattern, route := range r.routes {
		if route.WikiName == wikiName {
			delete(r.routes, pattern)
		}
	}

	// remove from static map (wiki routes are typically dynamic, but just in case)
	for pattern := range r.static {
		if strings.Contains(pattern, "/"+wikiName+"/") {
			delete(r.static, pattern)
		}
	}
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := req.URL.Path

	// try exact match first (fastest)
	if handler, exists := r.static[path]; exists {
		handler.ServeHTTP(w, req)
		return
	}

	// collect all matching routes and find the most specific one
	var bestMatch *Route
	bestSpecificity := -1

	for pattern, route := range r.routes {
		if params := r.matchPattern(pattern, path); params != nil {
			// calculate specificity: longer patterns and exact matches are more specific
			specificity := len(pattern)
			if !strings.HasSuffix(pattern, "/") && !strings.HasSuffix(pattern, "/*") {
				specificity += 1000 // exact matches are most specific
			}

			if specificity > bestSpecificity {
				bestMatch = route
				bestSpecificity = specificity
			}
		}
	}

	if bestMatch != nil {
		bestMatch.Handler.ServeHTTP(w, req)
		return
	}

	// no route found
	http.NotFound(w, req)
}

// matchPattern checks if a path matches a pattern and extracts parameters.
// supports simple wildcards like "/wiki/pages/*" and prefix routes like "/static/"
func (r *Router) matchPattern(pattern, path string) *Params {
	// prefix matching for routes ending with /
	if strings.HasSuffix(pattern, "/") {
		if strings.HasPrefix(path, pattern) {
			return &Params{} // no named params for prefix routes
		}
	}

	// simple wildcard matching for routes ending with /*
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(path, prefix+"/") || path == prefix {
			return &Params{} // no named params for simple wildcards
		}
	}

	// exact match
	if pattern == path {
		return &Params{}
	}

	return nil
}

// Routes returns a list of all registered routes for debugging.
func (r *Router) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes []Route

	// add static routes
	for pattern := range r.static {
		routes = append(routes, Route{
			Pattern:     pattern,
			Description: "static route",
		})
	}

	// add dynamic routes
	for _, route := range r.routes {
		routes = append(routes, *route)
	}

	return routes
}
