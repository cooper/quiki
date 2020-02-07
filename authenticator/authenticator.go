// Package authenticator provides server and site authentication services.
package authenticator

// Authenticator represents a quiki server or site authentication service.
type Authenticator struct {
}

// Open reads a user data file and returns an Authenticator for it.
func Open(path string) (*Authenticator, error) {
	return nil, nil
}
