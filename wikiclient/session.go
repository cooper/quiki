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

	// session ID used for write reauthentication
	// perhaps we can generate this automatically
	sessionID string

	ReadAccess  bool // true if authenticated for reading
	WriteAccess bool // true if authenticated for writing
}
