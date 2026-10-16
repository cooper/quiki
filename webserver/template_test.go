package webserver

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"

	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/router"
)

// the login and register pages share one template set, so they must not
// define templates with the same names: whichever parsed last would win
// and the login page would show the registration form
func TestAuthPagesRenderTheirOwnForms(t *testing.T) {
	Router = router.New()
	templates = make(map[string]wikiTemplate)

	themes, err := fs.Sub(resources.Webserver, "templates")
	if err != nil {
		t.Fatal(err)
	}
	sharedTemplateFs, err = fs.Sub(resources.Shared, "template")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := loadTemplate("default", "default", []fs.FS{themes})
	if err != nil {
		t.Fatal(err)
	}

	data := authTemplateData{
		WikiName:       "wiki",
		Title:          "Login - wiki",
		LoginAction:    "/w/login?redirect=%2Fw%2Fmain",
		RegisterAction: "/w/register",
		RegisterURL:    "/w/register",
		LoginURL:       "/w/login",
		HomeURL:        "/w",
		AllowRegister:  true,
	}

	var buf bytes.Buffer
	if err := tmpl.template.ExecuteTemplate(&buf, "login.tpl", data); err != nil {
		t.Fatal(err)
	}
	login := buf.String()
	for _, want := range []string{`action="/w/login?redirect=%2Fw%2Fmain"`, `name="password"`, `<title>Login - wiki</title>`, "Create account"} {
		if !strings.Contains(login, want) {
			t.Errorf("login page missing %q:\n%s", want, login)
		}
	}
	if strings.Contains(login, "password_confirm") {
		t.Errorf("login page shows the registration form:\n%s", login)
	}

	buf.Reset()
	if err := tmpl.template.ExecuteTemplate(&buf, "register.tpl", data); err != nil {
		t.Fatal(err)
	}
	register := buf.String()
	for _, want := range []string{`action="/w/register"`, `name="password_confirm"`, "already have an account?"} {
		if !strings.Contains(register, want) {
			t.Errorf("register page missing %q:\n%s", want, register)
		}
	}
}
