// Package authenticator provides server and site authentication services.
package authenticator

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/cooper/quiki/lock"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrInvalidUsername    = errors.New("invalid username")
)

// Authenticator represents a quiki server or site authentication service.
type Authenticator struct {
	Users    map[string]User `json:"users,omitempty"`
	Roles    map[string]Role `json:"roles,omitempty"`
	IsNew    bool            `json:"-"`
	IsServer bool            `json:"-"`
	path     string          // path to JSON file
	lock     *lock.Lock      // file lock
}

// Open reads a user file and returns an Authenticator for it.
// If the path does not exist, a new file is created.
func Open(path string) (*Authenticator, error) {
	return open(path, false)
}

// OpenServer reads a server auth file and returns an Authenticator for it.
// If the path does not exist, a new file is created.
func OpenServer(path string) (*Authenticator, error) {
	return open(path, true)
}

func open(path string, isServer bool) (*Authenticator, error) {
	auth := &Authenticator{
		path:     path,
		lock:     lock.New(path + ".lock"),
		IsServer: isServer,
	}

	// attempt to read the file
	jsonData, err := os.ReadFile(path)

	// it exists; try to unmarshal it
	if err == nil {
		err = json.Unmarshal(jsonData, auth)

		// JSON data is no good?
		// I mean, we can't just purge it because the data would be lost.
		// guess it needs some manual intervention...
		if err != nil {
			return nil, err
		}

		// all good
		return auth, nil
	}

	// hmm, a ReadFile error occurred OTHER THAN file does not exist
	if !os.IsNotExist(err) {
		return nil, err
	}

	// create a new one
	auth.IsNew = true

	// don't write anything to disk yet - file gets created when users are added
	return auth, nil
}

func (auth *Authenticator) write() error {
	// don't create file if there are no users and no custom roles
	if len(auth.Users) == 0 && len(auth.Roles) == 0 {
		return nil // nothing to persist
	}

	return auth.lock.WithLock(func() error {
		// encode as JSON
		jsonData, err := json.Marshal(auth)
		if err != nil {
			return err
		}

		// write
		return os.WriteFile(auth.path, jsonData, 0666)
	})
}

// Write overwrites the file with the current contents of the Authenticator.
func (auth *Authenticator) Write() error {
	return auth.write()
}

// MapUser creates a mapping between a server user and a wiki username
func (auth *Authenticator) MapUser(serverUser, wikiName, wikiUsername string) error {

	user, exists := auth.Users[serverUser]
	if !exists {
		return ErrUserNotFound
	}

	if user.Wikis == nil {
		user.Wikis = make(map[string]string)
	}

	user.Wikis[wikiName] = wikiUsername
	auth.Users[serverUser] = user
	return auth.write()
}

// GetWikiUsername returns the wiki username for a server user in a specific wiki
func (auth *Authenticator) GetWikiUsername(serverUser, wikiName string) (string, bool) {

	user, exists := auth.Users[serverUser]
	if !exists || user.Wikis == nil {
		return "", false
	}

	username, exists := user.Wikis[wikiName]
	return username, exists
}

// UnmapUser removes a server user's mapping to a wiki
func (auth *Authenticator) UnmapUser(serverUser, wikiName string) error {

	user, exists := auth.Users[serverUser]
	if !exists || user.Wikis == nil {
		return nil // nothing to remove
	}

	delete(user.Wikis, wikiName)
	auth.Users[serverUser] = user
	return auth.write()
}

// GetUserWikis returns all wikis that a server user is mapped to
func (auth *Authenticator) GetUserWikis(serverUser string) []string {

	user, exists := auth.Users[serverUser]
	if !exists || user.Wikis == nil {
		return nil
	}

	var wikis []string
	for wikiName := range user.Wikis {
		wikis = append(wikis, wikiName)
	}

	return wikis
}

// CreateUser creates a new user with the given username and password
func (auth *Authenticator) CreateUser(username, password string) error {
	// validate username
	if err := validateUsername(username); err != nil {
		return err
	}

	// validate password strength
	if err := validatePassword(password); err != nil {
		return err
	}

	if auth.Users == nil {
		auth.Users = make(map[string]User)
	}

	// check if user already exists
	if _, exists := auth.Users[username]; exists {
		return ErrUserExists
	}

	// hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	// create the user
	auth.Users[username] = User{
		Username:    username,
		Password:    hashedPassword,
		Roles:       []string{},
		Permissions: []string{},
		Wikis:       make(map[string]string),
	}

	return auth.write()
}

// GetUser returns a user by username
func (auth *Authenticator) GetUser(username string) (User, bool) {

	user, exists := auth.Users[username]
	return user, exists
}

// UpdateUser updates an existing user
func (auth *Authenticator) UpdateUser(username string, user User) error {

	if _, exists := auth.Users[username]; !exists {
		return ErrUserNotFound
	}

	auth.Users[username] = user
	return auth.write()
}

// DeleteUser removes a user
func (auth *Authenticator) DeleteUser(username string) error {

	if _, exists := auth.Users[username]; !exists {
		return ErrUserNotFound
	}

	delete(auth.Users, username)
	return auth.write()
}

// AddUserRole adds a role to a user
func (auth *Authenticator) AddUserRole(username, role string) error {

	user, exists := auth.Users[username]
	if !exists {
		return ErrUserNotFound
	}

	// check if user already has this role
	for _, r := range user.Roles {
		if r == role {
			return nil // already has the role
		}
	}

	user.Roles = append(user.Roles, role)
	auth.Users[username] = user
	return auth.write()
}

// RemoveUserRole removes a role from a user
func (auth *Authenticator) RemoveUserRole(username, role string) error {

	user, exists := auth.Users[username]
	if !exists {
		return ErrUserNotFound
	}

	// find and remove the role
	for i, r := range user.Roles {
		if r == role {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			auth.Users[username] = user
			return auth.write()
		}
	}

	return nil // role wasn't found, but that's ok
}

// ChangePassword changes a user's password
func (auth *Authenticator) ChangePassword(username, password string) error {

	user, exists := auth.Users[username]
	if !exists {
		return ErrUserNotFound
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	auth.Users[username] = user
	return auth.write()
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// GetAvailableRoles returns all available roles (built-in defaults + custom roles from file)
func (auth *Authenticator) GetAvailableRoles() map[string]Role {
	roles := make(map[string]Role)

	// start with built-in default roles for server auth
	if auth.IsServer {
		defaults := getDefaultServerRoles()
		for name, role := range defaults {
			roles[name] = role
		}
	}

	// add custom roles from file (these can override defaults)
	if auth.Roles != nil {
		for name, role := range auth.Roles {
			roles[name] = role
		}
	}

	return roles
}

// getDefaultServerRoles returns the built-in default roles for server authentication
// these are always available and get updated when the software is upgraded
func getDefaultServerRoles() map[string]Role {
	return map[string]Role{
		"admin": {
			Name:        "admin",
			Description: "full server administration access",
			Permissions: []string{"read.*", "write.*"},
		},
		"wiki-admin": {
			Name:        "wiki-admin",
			Description: "wiki administration access",
			Permissions: []string{"read.wiki.*", "write.wiki.*", "read.server.wikis"},
		},
		"editor": {
			Name:        "editor",
			Description: "content editing access",
			Permissions: []string{"read.wiki", "write.wiki.pages", "write.wiki.images"},
		},
		"viewer": {
			Name:        "viewer",
			Description: "read-only access",
			Permissions: []string{"read.wiki"},
		},
	}
}

// validatePassword checks if a password meets minimum security requirements
// returns a generic error to avoid leaking password policy details to attackers
func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password)
	)

	// require at least 3 out of 4 character types for passwords longer than 12 chars
	// or all 4 for shorter passwords
	requiredTypes := 4
	if len(password) >= 12 {
		requiredTypes = 3
	}

	foundTypes := 0
	if hasUpper {
		foundTypes++
	}
	if hasLower {
		foundTypes++
	}
	if hasNumber {
		foundTypes++
	}
	if hasSpecial {
		foundTypes++
	}

	if foundTypes < requiredTypes {
		return ErrWeakPassword
	}

	// check for common weak patterns (case insensitive)
	lowerPassword := strings.ToLower(password)
	weakPatterns := []string{
		"password", "123456", "qwerty", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "trustno1",
		"iloveyou", "princess", "1234567890", "abc123",
		"login", "user", "test", "guest", "root", "administrator",
		"passw0rd", "p@ssword", "password123", "123password",
	}

	for _, pattern := range weakPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return ErrWeakPassword
		}
	}

	// check for sequential characters (like "12345", "abcde")
	if hasSequentialChars(password) {
		return ErrWeakPassword
	}

	return nil
}

// hasSequentialChars checks for sequential characters in password
func hasSequentialChars(password string) bool {
	if len(password) < 4 {
		return false
	}

	lower := strings.ToLower(password)
	for i := 0; i < len(lower)-3; i++ {
		// check for sequential numbers
		if lower[i] >= '0' && lower[i] <= '6' {
			sequential := true
			for j := 1; j < 4; j++ {
				if lower[i+j] != lower[i]+byte(j) {
					sequential = false
					break
				}
			}
			if sequential {
				return true
			}
		}

		// check for sequential letters
		if lower[i] >= 'a' && lower[i] <= 'w' {
			sequential := true
			for j := 1; j < 4; j++ {
				if lower[i+j] != lower[i]+byte(j) {
					sequential = false
					break
				}
			}
			if sequential {
				return true
			}
		}
	}

	return false
}

// validateUsername checks if a username is valid
func validateUsername(username string) error {
	if len(username) == 0 {
		return ErrInvalidUsername
	}

	if len(username) < 2 || len(username) > 32 {
		return ErrInvalidUsername
	}

	// only allow alphanumeric, underscore, and hyphen
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return ErrInvalidUsername
	}

	// can't start or end with special chars
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, "-") {
		return ErrInvalidUsername
	}

	return nil
}
