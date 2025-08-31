package webserver

import (
	"context"
	"net/http"

	"github.com/cooper/quiki/authenticator"
)

// SessionManager interface for permission caching
type SessionManager interface {
	Get(ctx context.Context, key string) interface{}
	Put(ctx context.Context, key string, val interface{})
}

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
	user, ok := pc.sessionMgr.Get(r.Context(), "user").(*authenticator.User)
	if !ok || user == nil {
		return false
	}

	// check cached permissions first
	cacheKey := "server_perms:" + user.Username
	if cachedPerms, ok := pc.sessionMgr.Get(r.Context(), cacheKey).(map[string]bool); ok {
		if result, exists := cachedPerms[required]; exists {
			return result
		}
	}

	// expand user roles to get all permissions
	availableRoles := Auth.GetAvailableRoles()
	allPermissions := authenticator.ExpandRolePermissions(user.Roles, availableRoles)
	allPermissions = append(allPermissions, user.Permissions...)

	// check permission
	result := authenticator.CheckPermission(allPermissions, required)

	// cache the result
	var cachedPerms map[string]bool
	if existing, ok := pc.sessionMgr.Get(r.Context(), cacheKey).(map[string]bool); ok {
		cachedPerms = existing
	} else {
		cachedPerms = make(map[string]bool)
	}
	cachedPerms[required] = result
	pc.sessionMgr.Put(r.Context(), cacheKey, cachedPerms)

	return result
}

// HasWikiPermission checks if the current user has a specific wiki permission with caching
func (pc *PermissionChecker) HasWikiPermission(r *http.Request, wikiName, required string) bool {
	user, ok := pc.sessionMgr.Get(r.Context(), "user").(*authenticator.User)
	if !ok || user == nil {
		return false
	}

	// check cached permissions first
	cacheKey := "wiki_perms:" + user.Username + ":" + wikiName
	if cachedPerms, ok := pc.sessionMgr.Get(r.Context(), cacheKey).(map[string]bool); ok {
		if result, exists := cachedPerms[required]; exists {
			return result
		}
	}

	// first check if user has server-level permissions that would grant this wiki permission
	// e.g., read.* would grant read.wiki.*, write.* would grant write.wiki.*
	availableRoles := Auth.GetAvailableRoles()
	serverPerms := authenticator.ExpandRolePermissions(user.Roles, availableRoles)
	serverPerms = append(serverPerms, user.Permissions...)
	if authenticator.CheckPermission(serverPerms, required) {
		// cache the positive result
		pc.cacheWikiPermission(r.Context(), cacheKey, required, true)
		return true
	}

	// no server-level permission, check if user is mapped to this wiki
	wikiUsername, hasMappings := Auth.GetWikiUsername(user.Username, wikiName)
	if !hasMappings {
		// cache negative result
		pc.cacheWikiPermission(r.Context(), cacheKey, required, false)
		return false
	}

	// get the wiki info
	wikiInfo, exists := Wikis[wikiName]
	if !exists {
		// cache negative result
		pc.cacheWikiPermission(r.Context(), cacheKey, required, false)
		return false
	}

	// get the wiki's auth to check permissions
	wikiAuthPath := wikiInfo.Wiki.Dir("auth.json")
	wikiAuth, err := authenticator.Open(wikiAuthPath)
	if err != nil {
		// cache negative result
		pc.cacheWikiPermission(r.Context(), cacheKey, required, false)
		return false
	}

	wikiUser, exists := wikiAuth.GetUser(wikiUsername)
	if !exists {
		// cache negative result
		pc.cacheWikiPermission(r.Context(), cacheKey, required, false)
		return false
	}

	// check permission using wiki roles
	wikiRoles := wikiAuth.GetAvailableRoles()
	result := wikiUser.HasPermission(required, wikiRoles)

	// cache the result
	pc.cacheWikiPermission(r.Context(), cacheKey, required, result)

	return result
}

func (pc *PermissionChecker) cacheWikiPermission(ctx context.Context, cacheKey, required string, result bool) {
	var cachedPerms map[string]bool
	if existing, ok := pc.sessionMgr.Get(ctx, cacheKey).(map[string]bool); ok {
		cachedPerms = existing
	} else {
		cachedPerms = make(map[string]bool)
	}
	cachedPerms[required] = result
	pc.sessionMgr.Put(ctx, cacheKey, cachedPerms)
}

// ClearPermissionCache clears cached permissions for a user
func (pc *PermissionChecker) ClearPermissionCache(r *http.Request, username string) {
	ctx := r.Context()

	// clear server permissions cache
	serverCacheKey := "server_perms:" + username
	pc.sessionMgr.Put(ctx, serverCacheKey, nil)

	// clear wiki permissions cache for all wikis
	for wikiName := range Wikis {
		wikiCacheKey := "wiki_perms:" + username + ":" + wikiName
		pc.sessionMgr.Put(ctx, wikiCacheKey, nil)
	}
}
