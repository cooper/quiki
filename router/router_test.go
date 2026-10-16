package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// a handler must be able to register routes while it is running. the server
// rehashes (and so registers wiki routes) from inside request handlers, which
// deadlocked when ServeHTTP held the router lock for the whole request.
func TestHandlerCanRegisterRoutes(t *testing.T) {
	r := New()
	r.HandleFunc("/admin/func/create-wiki", "create wiki", func(w http.ResponseWriter, req *http.Request) {
		r.HandleFunc("/newwiki/", "new wiki", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		w.WriteHeader(http.StatusSeeOther)
	})

	done := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest("POST", "/admin/func/create-wiki", nil))
		done <- rec.Code
	}()

	select {
	case code := <-done:
		if code != http.StatusSeeOther {
			t.Fatalf("create-wiki handler returned %d", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: handler blocked registering a route while being served")
	}

	// the route registered from inside the handler is now served
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/newwiki/page", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("new route not served, got %d", rec.Code)
	}
}

// most specific route wins, and unknown paths 404
func TestRouting(t *testing.T) {
	r := New()
	hit := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) { w.Header().Set("X-Route", name) }
	}
	r.HandleFunc("/admin/", "admin prefix", hit("admin"))
	r.HandleFunc("/admin/login", "admin login", hit("login"))
	r.HandleFunc("/admin/-/w/frame/*", "frames", hit("frames"))
	r.HandleFunc("example.com/", "host root", hit("host"))

	tests := []struct{ host, path, want string }{
		{"", "/admin/", "admin"},
		{"", "/admin", "admin"},
		{"", "/admin/sites", "admin"},
		{"", "/admin/login", "login"},
		{"", "/admin/-/w/frame/pages/sub", "frames"},
		{"example.com", "/anything", "host"},
		{"example.com:8080", "/", "host"},
		{"other.com", "/nothing", ""},
	}
	for _, tt := range tests {
		// httptest defaults Host to example.com, which would match the host route
		req := httptest.NewRequest("GET", tt.path, nil)
		req.Host = "localhost"
		if tt.host != "" {
			req.Host = tt.host
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if got := rec.Header().Get("X-Route"); got != tt.want {
			t.Errorf("%s%s: routed to %q, want %q (status %d)", tt.host, tt.path, got, tt.want, rec.Code)
		}
		if tt.want == "" && rec.Code != http.StatusNotFound {
			t.Errorf("%s%s: status %d, want 404", tt.host, tt.path, rec.Code)
		}
	}
}
