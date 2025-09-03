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
	Host        string
	Path        string
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

	host, path := parsePattern(pattern)

	route := &Route{
		Pattern:     pattern,
		Host:        host,
		Path:        path,
		Description: description,
		Handler:     handler,
	}

	// if pattern has no wildcards AND doesn't end with /, treat as static route for fast lookup
	// but only if there's no host component
	if host == "" && !strings.Contains(path, "*") && !strings.Contains(path, ":") && !strings.HasSuffix(path, "/") {
		r.static[path] = handler
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
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reqPath := req.URL.Path

	// try exact match first (fastest)
	if handler, exists := r.static[reqPath]; exists {
		handler.ServeHTTP(w, req)
		return
	}

	// collect all matching routes and find the most specific one
	var bestMatch *Route
	bestSpecificity := -1

	reqHost := req.Host

	for _, route := range r.routes {
		if r.matchRoute(route, reqHost, reqPath) {
			// calculate specificity: longer patterns and exact matches are more specific
			specificity := len(route.Path)
			if route.Host != "" {
				// host-specific routes are more specific
				specificity += len(route.Host) + 100
			}
			if !strings.HasSuffix(route.Path, "/") && !strings.HasSuffix(route.Path, "/*") {
				// exact matches are super specific
				specificity += 1000
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
		// exact match (e.g., "/admin/" matches "/admin/")
		if path == pattern {
			return &Params{}
		}
		// prefix match (e.g., "/admin/" matches "/admin/something")
		if strings.HasPrefix(path, pattern) {
			return &Params{}
		}
		// also match without trailing slash (e.g., "/admin/" matches "/admin")
		if path == strings.TrimSuffix(pattern, "/") {
			return &Params{}
		}
	}

	// simple wildcard matching for routes ending with /*
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(path, prefix+"/") || path == prefix {
			return &Params{} // no named params for simple wildcards
		}
	}

	// prefix routes ending with / should also match subpaths
	if strings.HasSuffix(pattern, "/") {
		if strings.HasPrefix(path, pattern) {
			return &Params{}
		}
	}

	// exact match
	if pattern == path {
		return &Params{}
	}

	return nil
}

func (r *Router) matchRoute(route *Route, reqHost, reqPath string) bool {
	// check host match first
	if route.Host != "" {
		// strip port
		host := reqHost
		if idx := strings.Index(reqHost, ":"); idx != -1 {
			host = reqHost[:idx]
		}

		if route.Host != host {
			return false
		}
	}

	// check path match
	return r.matchPattern(route.Path, reqPath) != nil
}

// parsePattern splits a pattern into host and path components.
// e.g. "example.com/api/" -> ("example.com", "/api/")
// e.g. "/api/" -> ("", "/api/")
func parsePattern(pattern string) (host, path string) {
	// strip scheme
	if strings.Contains(pattern, "://") {
		if idx := strings.Index(pattern, "://"); idx != -1 {
			pattern = pattern[idx+3:]
		}
	}

	if idx := strings.Index(pattern, "/"); idx != -1 {
		// host with path
		host = pattern[:idx]
		path = pattern[idx:]

		if host == "" || strings.Contains(host, ":") {
			host = ""
			path = pattern
		}
	} else {
		// no slash, treat as path
		host = ""
		path = pattern
	}

	return host, path
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
