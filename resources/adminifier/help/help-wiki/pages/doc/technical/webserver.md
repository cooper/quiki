# webserver
--
    import "."

Package webserver is the newest webserver.

## Usage

```go
var Auth *authenticator.Authenticator
```
Auth is the server authentication service.

```go
var Bind string
```
Bind is the string to bind to, as extracted from the configuration file.

It is available only after Configure is called.

```go
var Conf *wikifier.Page
```
Conf is the webserver configuration page.

It is available only after Configure is called.

```go
var Mux *http.ServeMux
```
Mux is the *http.ServeMux.

It is available only after Configure is called.

```go
var Port string
```
Port is the port to bind to or "unix" for a UNIX socket, as extracted from the
configuration file.

It is available only after Configure is called.

```go
var Server *http.Server
```
Server is the *http.Server.

It is available only after Configure is called.

```go
var SessMgr *scs.SessionManager
```
SessMgr is the session manager service.

```go
var Wikis map[string]*WikiInfo
```
Wikis is all wikis served by this webserver.

#### func  Configure

```go
func Configure(opts Options)
```
Configure parses a configuration file and initializes webserver.

If any errors occur, the program is terminated.

#### func  CreateWizardConfig

```go
func CreateWizardConfig(opts Options)
```

#### func  InitWikis

```go
func InitWikis() error
```
initialize all the wikis in the configuration

#### func  Listen

```go
func Listen()
```
Listen runs the webserver indefinitely.

Configure must be called first. If any errors occur, the program is terminated.

#### func  TemplateNames

```go
func TemplateNames() []string
```
Returns the names of all available templates.

#### type Options

```go
type Options struct {
	Config string
	Bind   string
	Port   string
	Pregen bool
}
```

Options is the webserver command line options.

#### type WikiInfo

```go
type WikiInfo struct {
	Name  string // wiki shortname
	Title string // wiki title from @name in the wiki config
	Logo  string
	Host  string

	*wiki.Wiki
}
```

WikiInfo represents a wiki hosted on this webserver.

#### func (*WikiInfo) Copy

```go
func (wi *WikiInfo) Copy(w *wiki.Wiki) *WikiInfo
```
Copy creates a WikiInfo with all the same options, minus Wiki. It is used for
working with multiple branches within a wiki.
