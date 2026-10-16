package webserver

import (
	"testing"

	"github.com/cooper/quiki/wiki"
)

func wikiWithRoot(root string) *WikiInfo {
	wi := &WikiInfo{Wiki: &wiki.Wiki{}}
	wi.Opt.Root.Wiki = root
	return wi
}

// auth URLs must be absolute paths under the wiki root, whether the wiki
// lives at the host root (Root.Wiki is empty) or under a prefix
func TestAuthURLs(t *testing.T) {
	tests := []struct {
		root, home, login string
	}{
		{"", "/", "/login"},
		{"/", "/", "/login"},
		{"/mywiki", "/mywiki", "/mywiki/login"},
		{"/mywiki/", "/mywiki", "/mywiki/login"},
		{"/a/b", "/a/b", "/a/b/login"},
	}
	for _, tt := range tests {
		wi := wikiWithRoot(tt.root)
		if got := wi.homeURL(); got != tt.home {
			t.Errorf("root %q: homeURL() = %q, want %q", tt.root, got, tt.home)
		}
		if got := wi.authURL("login"); got != tt.login {
			t.Errorf("root %q: authURL(login) = %q, want %q", tt.root, got, tt.login)
		}
	}
}

// only local paths may be used as a post-login redirect
func TestSafeRedirect(t *testing.T) {
	tests := []struct {
		target, want string
	}{
		{"", "/home"},
		{"/", "/"},
		{"/page?x=y", "/page?x=y"},
		{"//evil.example", "/home"},
		{"/\\evil.example", "/home"},
		{"https://evil.example/", "/home"},
		{"page", "/home"},
		{"/\t/evil.example", "/home"},
		{"/\t\t/evil.example", "/home"},
		{"/\r\n/evil.example", "/home"},
		{"/%09/evil.example", "/%09/evil.example"},
		{"/page%20name?x=a%26b#frag", "/page%20name?x=a%26b#frag"},
	}
	for _, tt := range tests {
		if got := safeRedirect(tt.target, "/home"); got != tt.want {
			t.Errorf("safeRedirect(%q) = %q, want %q", tt.target, got, tt.want)
		}
	}
}
