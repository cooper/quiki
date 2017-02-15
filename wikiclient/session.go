// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// a Session represents a user session.
//
// for read-only authentication, a single session may be used for all requests.
//
// for write authentication, there should exist one session per user. the
// program should retain the session instance as long as the session exists. it
// should implement some sort of session timeout disposal.
//

type Session struct {

	// wiki credentials for read authentication
	WikiName     string
	WikiPassword string

	// optional user credentials for write authentication
	UserName     string
	UserPassword string

	ReadAccess  bool // true if authenticated for reading
	WriteAccess bool // true if authenticated for writing

	// session ID used for write reauthentication
	// perhaps we can generate this automatically
	sessionID string

	// ID of the transport at the last clean
	transportID uint
}

// prepares the session for use with the given transport
func (sess *Session) Clean(tr Transport) {

	// nothing has changed
	if sess.transportID == tr.ID() {
		return
	}

	sess.transportID = tr.ID()
	sess.ReadAccess = false
	sess.WriteAccess = false
}
