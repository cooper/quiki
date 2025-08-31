package authenticator

import (
	"strings"
)

// Permission represents a permission to a single action
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Role represents a collection of permissions
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Inherits    []string `json:"inherits,omitempty"` // roles this role inherits from
}

// CheckPermission checks if a set of user permissions matches a required permission
func CheckPermission(userPermissions []string, required string) bool {
	for _, perm := range userPermissions {
		if permissionMatches(perm, required) {
			return true
		}
	}
	return false
}

// permissionMatches checks if a granted permission matches a requirement.
// supports wildcards like "read.*" matching "read.page.somepage"
func permissionMatches(granted, required string) bool {
	// exact match
	if granted == required {
		return true
	}

	// wildcard match
	if strings.HasSuffix(granted, ".*") {
		prefix := strings.TrimSuffix(granted, ".*")
		return strings.HasPrefix(required, prefix+".")
	}

	// catch-all wildcard
	if granted == "*" {
		return true
	}

	return false
}

// ExpandRolePermissions takes roles and returns all permissions including inherited ones
func ExpandRolePermissions(userRoles []string, availableRoles map[string]Role) []string {
	seen := make(map[string]bool)
	var permissions []string

	var expandRole func(roleName string)
	expandRole = func(roleName string) {
		if seen[roleName] {
			return // avoid infinite loops
		}
		seen[roleName] = true

		role, exists := availableRoles[roleName]
		if !exists {
			return
		}

		// first expand inherited roles
		for _, inheritedRole := range role.Inherits {
			expandRole(inheritedRole)
		}

		// then add this role's permissions
		permissions = append(permissions, role.Permissions...)
	}

	for _, roleName := range userRoles {
		expandRole(roleName)
	}

	return permissions
}
