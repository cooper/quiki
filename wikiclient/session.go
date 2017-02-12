// Copyright (c) 2017, Mitchell Cooper
package wikiclient

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
}
