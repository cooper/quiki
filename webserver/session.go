package webserver

import (
	"context"

	"github.com/cooper/quiki/authenticator"
)

// Session embeds the authenticator.User and adds permission caching
type Session struct {
	authenticator.User
	ServerPermissions map[string]bool            `json:"server_perms,omitempty"`
	WikiPermissions   map[string]map[string]bool `json:"wiki_perms,omitempty"`
}

// NewSession creates a new session user from an authenticator user
func NewSession(user *authenticator.User) *Session {
	return (&Session{User: *user}).init()
}

func (s *Session) init() *Session {
	if s.ServerPermissions == nil {
		s.ServerPermissions = make(map[string]bool)
	}
	if s.WikiPermissions == nil {
		s.WikiPermissions = make(map[string]map[string]bool)
	}
	return s
}

// SessionManager interface for permission caching
type SessionManager interface {
	Get(ctx context.Context, key string) interface{}
	Put(ctx context.Context, key string, val interface{})
}
