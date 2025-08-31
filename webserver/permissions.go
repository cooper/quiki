package webserver

import (
	"net/http"

	"github.com/cooper/quiki/authenticator"
)

// PermissionChecker provides cached permission checking for web requests
type PermissionChecker struct {
	sessionMgr SessionManager
}

// NewPermissionChecker creates a new permission checker with the given session manager
func NewPermissionChecker(sessionMgr SessionManager) *PermissionChecker {
	return &PermissionChecker{sessionMgr: sessionMgr}
}

// HasServerPermission checks if the current user has a specific server permission with caching
func (pc *PermissionChecker) HasServerPermission(r *http.Request, required string) bool {
	session, ok := pc.sessionMgr.Get(r.Context(), "user").(*Session)
	if !ok || session == nil || session.Username == "" {
		return false
	}

	session.init()

	// check cached permissions first
	if result, exists := session.ServerPermissions[required]; exists {
		return result
	}

	// expand user roles to get all permissions
	availableRoles := Auth.GetAvailableRoles()
	allPermissions := authenticator.ExpandRolePermissions(session.Roles, availableRoles)
	allPermissions = append(allPermissions, session.Permissions...)

	// check permission
	result := authenticator.CheckPermission(allPermissions, required)

	// cache the result
	session.ServerPermissions[required] = result
	pc.sessionMgr.Put(r.Context(), "user", session)

	return result
}

// HasWikiPermission checks if the current user has a specific wiki permission with caching
func (pc *PermissionChecker) HasWikiPermission(r *http.Request, wikiName, required string) bool {
	session, ok := pc.sessionMgr.Get(r.Context(), "user").(*Session)
	if !ok || session == nil || session.Username == "" {
		return false
	}

	// ensure maps are initialized (for deserialized sessions)
	session.init()

	// check cached permissions first
	if wikiPerms, exists := session.WikiPermissions[wikiName]; exists {
		if result, exists := wikiPerms[required]; exists {
			return result
		}
	}

	// first check if user has server-level permissions that would grant this wiki permission
	if pc.HasServerPermission(r, required) {
		return true
	}

	// get wiki-specific permissions
	allPermissions := authenticator.ExpandRolePermissions(session.Roles, Auth.GetAvailableRoles())
	allPermissions = append(allPermissions, session.Permissions...)

	// check permission
	result := authenticator.CheckPermission(allPermissions, required)

	// cache the result
	if session.WikiPermissions[wikiName] == nil {
		session.WikiPermissions[wikiName] = make(map[string]bool)
	}
	session.WikiPermissions[wikiName][required] = result
	pc.sessionMgr.Put(r.Context(), "user", session)

	return result
}

// ClearPermissionCache clears cached permissions for a user
func (pc *PermissionChecker) ClearPermissionCache(r *http.Request, username string) {
	session, ok := pc.sessionMgr.Get(r.Context(), "user").(*Session)
	if !ok || session == nil {
		return
	}

	// clear all cached permissions
	session.ServerPermissions = make(map[string]bool)
	session.WikiPermissions = make(map[string]map[string]bool)
	pc.sessionMgr.Put(r.Context(), "user", session)
}
