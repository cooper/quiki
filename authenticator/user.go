package authenticator

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user.
type User struct {
	Username    string `json:"u"`
	DisplayName string `json:"d"`
	Email       string `json:"e"`
	Password    []byte `json:"p"`
}

// NewUser registers a new user with the given information.
//
// The Password field of the struct should be left empty and
// the plain-text password passed to the function.
//
func (auth *Authenticator) NewUser(user User, password string) error {
	// consider: is it possible 2 users could be created with the same username
	// at the same time?
	lcun := strings.ToLower(user.Username)

	// user already exists!!
	if _, exist := auth.Users[lcun]; exist {
		return errors.New("user exists")
	}

	// hash password
	var err error
	user.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// store the user
	auth.Users[lcun] = user

	// write to file
	return auth.write()
}
