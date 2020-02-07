package authenticator

import (
	"encoding/json"
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
	if auth.Users == nil {
		auth.Users = make(map[string]User)
	}
	auth.Users[lcun] = user

	// write to file
	return auth.write()
}

// Login attempts a user login, returning the user on success.
//
func (auth *Authenticator) Login(username, password string) (User, error) {
	lcun := strings.ToLower(username)

	// user does not exist
	user, exist := auth.Users[lcun]
	if !exist {
		return user, errors.New("user does not exist")
	}

	// bad password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return user, errors.New("bad password")
	}

	return user, nil
}

// GobDecode allows users to be decoded from a session.
func (user *User) GobDecode(data []byte) error {
	return json.Unmarshal(data, user)
}

// GobEncode allows users to be encoded for storage in a session.
func (user *User) GobEncode() ([]byte, error) {
	return json.Marshal(user)
}
