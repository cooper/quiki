# authenticator
--
    import "."

Package authenticator provides server and site authentication services.

## Usage

#### type Authenticator

```go
type Authenticator struct {
	Users map[string]User `json:"users,omitempty"`
	IsNew bool            `json:"-"`
}
```

Authenticator represents a quiki server or site authentication service.

#### func  Open

```go
func Open(path string) (*Authenticator, error)
```
Open reads a user data file and returns an Authenticator for it. If the path
does not exist, a new data file is created.

#### func (*Authenticator) Login

```go
func (auth *Authenticator) Login(username, password string) (User, error)
```
Login attempts a user login, returning the user on success.

#### func (*Authenticator) NewUser

```go
func (auth *Authenticator) NewUser(user User, password string) error
```
NewUser registers a new user with the given information.

The Password field of the struct should be left empty and the plain-text
password passed to the function.

#### type User

```go
type User struct {
	Username    string `json:"u"`
	DisplayName string `json:"d"`
	Email       string `json:"e"`
	Password    []byte `json:"p"`
}
```

User represents a user.

#### func (*User) GobDecode

```go
func (user *User) GobDecode(data []byte) error
```
GobDecode allows users to be decoded from a session.

#### func (*User) GobEncode

```go
func (user *User) GobEncode() ([]byte, error)
```
GobEncode allows users to be encoded for storage in a session.
