package wiki

import "github.com/cooper/quiki/wikifier"

// DisplayError represents an error result to display.
type DisplayError struct {
	// a human-readable error string. sensitive info is never
	// included, so this may be shown to users
	Error string

	// a more detailed human-readable error string that MAY contain
	// sensitive data. can be used for debugging and logging but should
	// not be presented to users
	DetailedError string

	// HTTP status code. if zero, 404 should be used
	Status int

	// if the error occurred during parsing, this is the position.
	// for all non-parsing errors, this is 0:0
	Pos wikifier.Position

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

func (e DisplayError) ErrorAsWarning() wikifier.Warning {
	return wikifier.Warning{
		Message: e.Error,
		Pos:     e.Pos,
	}
}
