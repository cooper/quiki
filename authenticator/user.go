package authenticator

import (
	"encoding/json"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user.
type User struct {
	Username    string   `json:"u"`
	DisplayName string   `json:"d"`
	Email       string   `json:"e"`
	Password    []byte   `json:"p"`
	Roles       []string `json:"r,omitempty"`
	Permissions []string `json:"a,omitempty"`

	// for server users, this maps wiki name to wiki username
	Wikis map[string]string `json:"w,omitempty"`
}

// Login attempts a user login, returning the user on success.
// uses generic error messages to prevent user enumeration attacks
func (auth *Authenticator) Login(username, password string) (User, error) {
	lcun := strings.ToLower(username)

	// check if user exists and password is correct
	user, exist := auth.Users[lcun]
	if !exist {
		return user, ErrInvalidCredentials // don't reveal that user doesn't exist
	}

	// verify password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return user, ErrInvalidCredentials // use same error as above
	}

	return user, nil
}

// HasPermission checks if user has a specific permission (including role permissions)
func (user *User) HasPermission(required string, availableRoles map[string]Role) bool {
	// check direct permissions first
	if CheckPermission(user.Permissions, required) {
		return true
	}

	// check role permissions
	rolePermissions := ExpandRolePermissions(user.Roles, availableRoles)
	return CheckPermission(rolePermissions, required)
}

// GobDecode allows users to be decoded from a session.
func (user *User) GobDecode(data []byte) error {
	return json.Unmarshal(data, user)
}

// GobEncode allows users to be encoded for storage in a session.
func (user *User) GobEncode() ([]byte, error) {
	return json.Marshal(user)
}
