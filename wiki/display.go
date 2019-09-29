package wiki

// DisplayError represents an error result to display.
type DisplayError struct {
	// a human-readable error string. sensitive info is never
	// included, so this may be shown to users
	Error string

	// a more detailed human-readable error string that MAY contain
	// sensitive data. can be used for debugging and logging but should
	// not be presented to users
	DetailedError string

	// true if the error occurred during parsing
	ParseError bool

	// true if the content cannot be displayed because it has
	// not yet been published for public access
	Draft bool
}

// DisplayRedirect represents a page redirect to follow.
type DisplayRedirect struct {

	// a relative or absolute URL to which the request should redirect,
	// suitable for use in a Location header
	Redirect string
}
