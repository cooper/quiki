// Package authenticator provides server and site authentication services.
package authenticator

import (
	"encoding/json"
	"os"
	"sync"
)

// Authenticator represents a quiki server or site authentication service.
type Authenticator struct {
	Users map[string]User `json:"users,omitempty"`
	IsNew bool            `json:"-"`
	path  string          // path to JSON file
	mu    *sync.Mutex     // data lock
}

// Open reads a user data file and returns an Authenticator for it.
// If the path does not exist, a new data file is created.
func Open(path string) (*Authenticator, error) {
	auth := &Authenticator{path: path, mu: new(sync.Mutex)}

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
	return auth, auth.write()
}

// Write overwrites the data file with the current contents of the Authenticator.
func (auth *Authenticator) write() error {
	auth.mu.Lock()
	defer auth.mu.Unlock()

	// encode as JSON
	jsonData, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	// write
	return os.WriteFile(auth.path, jsonData, 0666)
}
