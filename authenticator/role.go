package authenticator

// GetAvailableRoles returns all available roles (built-in defaults + custom roles from file)
func (auth *Authenticator) GetAvailableRoles() map[string]Role {
	roles := make(map[string]Role)

	// start with built-in default roles for server auth
	if auth.IsServer {
		defaults := getDefaultServerRoles()
		for name, role := range defaults {
			roles[name] = role
		}
	}

	// add custom roles from file (these can override defaults)
	if auth.Roles != nil {
		for name, role := range auth.Roles {
			roles[name] = role
		}
	}

	return roles
}

func getDefaultServerRoles() map[string]Role {
	return map[string]Role{
		"admin": {
			Name:        "admin",
			Description: "full server administration access",
			Permissions: []string{"read.*", "write.*"},
		},
		"wiki-admin": {
			Name:        "wiki-admin",
			Description: "wiki administration access",
			Permissions: []string{"read.wiki.*", "write.wiki.*", "read.server.wikis"},
		},
		"editor": {
			Name:        "editor",
			Description: "content editing access",
			Permissions: []string{"read.wiki", "write.wiki.pages", "write.wiki.images"},
		},
		"viewer": {
			Name:        "viewer",
			Description: "read-only access",
			Permissions: []string{"read.wiki"},
		},
	}
}
