package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cooper/quiki/authenticator"
	"golang.org/x/term"
)

// this file handles CLI commands for managing users and roles

func handleAuthCommand() {
	if len(os.Args) < 3 {
		printAuthUsage()
		return
	}

	var auth *authenticator.Authenticator
	var err error
	if wikiPath == "" {
		// server-level auth - use the established quiki directory
		var authPath string
		if QuikiDir != "" {
			authPath = filepath.Join(QuikiDir, "quiki-auth.json")
		} else {
			authPath = "quiki-auth.json" // fallback for relative path
		}
		auth, err = authenticator.OpenServer(authPath)
	} else {
		// wiki-level auth
		auth, err = authenticator.Open(wikiPath + "/auth.json")
	}

	if err != nil {
		log.Fatalf("error opening auth file: %v", err)
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "create-user":
		handleCreateUser(auth)
	case "delete-user":
		handleDeleteUser(auth)
	case "list-users":
		handleListUsers(auth)
	case "change-password":
		handleChangePassword(auth)
	case "add-role":
		handleAddRole(auth)
	case "remove-role":
		handleRemoveRole(auth)
	case "list-roles":
		handleListRoles(auth)
	case "map-user":
		handleMapUser(auth)
	case "unmap-user":
		handleUnmapUser(auth)
	case "list-mappings":
		handleListMappings(auth)
	default:
		fmt.Printf("unknown auth subcommand: %s\n", subcommand)
		printAuthUsage()
	}
}

func printAuthUsage() {
	fmt.Println("quiki auth management usage:")
	fmt.Println("")
	fmt.Println("server or wiki commands")
	fmt.Println("(optionally pass -wiki, otherwise operating on server users):")
	fmt.Println("  quiki auth create-user <username>     create a new user")
	fmt.Println("  quiki auth delete-user <username>     delete an existing user")
	fmt.Println("  quiki auth list-users                 list all users")
	fmt.Println("  quiki auth change-password <username> change user password")
	fmt.Println("  quiki auth add-role <username> <role> add role to user")
	fmt.Println("  quiki auth remove-role <username> <role> remove role from user")
	fmt.Println("  quiki auth list-roles                 list all available roles")
	fmt.Println("")
	fmt.Println("server-only commands")
	fmt.Println("(for assigning server users to wikis):")
	fmt.Println("  quiki auth map-user <server-user> <wiki-name> <wiki-username>")
	fmt.Println("                                        map server user to wiki user")
	fmt.Println("  quiki auth unmap-user <server-user> <wiki-name>")
	fmt.Println("                                        remove user mapping from wiki")
	fmt.Println("  quiki auth list-mappings [server-user]")
	fmt.Println("                                        list user mappings")
	fmt.Println("")
	fmt.Println("flags:")
	fmt.Println("  -wiki=/path/to/wiki                   operate on wiki auth")
	fmt.Println("                                        (when omitted, server auth)")
	fmt.Println("")
	fmt.Println("examples:")
	fmt.Println("  quiki auth create-user admin          create server admin user")
	fmt.Println("  quiki auth -wiki=/path/to/wiki create-user someone")
	fmt.Println("                                        create wiki-specific user")
	fmt.Println("  quiki auth add-role admin wiki-admin  give admin user wiki-admin role")
	fmt.Println("  quiki auth map-user cooper mywiki cooper-wiki")
	fmt.Println("                                        map server user to wiki username")
}

func handleCreateUser(auth *authenticator.Authenticator) {
	if len(os.Args) < 4 {
		fmt.Println("usage: quiki auth create-user <username>")
		return
	}

	username := os.Args[3]

	// check if user already exists
	if _, exists := auth.Users[username]; exists {
		fmt.Printf("user %s already exists\n", username)
		return
	}

	// get password
	fmt.Print("password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("error reading password: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Print("confirm password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("error reading password: %v\n", err)
		return
	}
	fmt.Println()

	if string(password) != string(confirmPassword) {
		fmt.Println("passwords do not match")
		return
	}

	// create user
	_, err = auth.CreateUser(username, string(password))
	if err != nil {
		fmt.Printf("error creating user: %v\n", err)
		return
	}

	fmt.Printf("user %s created successfully\n", username)
}

func handleDeleteUser(auth *authenticator.Authenticator) {
	if len(os.Args) < 4 {
		fmt.Println("usage: quiki auth delete-user <username>")
		return
	}

	username := os.Args[3]

	if _, exists := auth.Users[username]; !exists {
		fmt.Printf("user %s does not exist\n", username)
		return
	}

	err := auth.DeleteUser(username)
	if err != nil {
		fmt.Printf("error deleting user: %v\n", err)
		return
	}

	fmt.Printf("user %s deleted successfully\n", username)
}

func handleListUsers(auth *authenticator.Authenticator) {
	if len(auth.Users) == 0 {
		fmt.Println("no users found")
		return
	}

	fmt.Println("users:")
	for username, user := range auth.Users {
		fmt.Printf("  %s", username)
		if len(user.Roles) > 0 {
			fmt.Printf(" (roles: %v)", user.Roles)
		}
		if len(user.Permissions) > 0 {
			fmt.Printf(" (permissions: %v)", user.Permissions)
		}
		fmt.Println()
	}
}

func handleChangePassword(auth *authenticator.Authenticator) {
	if len(os.Args) < 4 {
		fmt.Println("usage: quiki auth change-password <username>")
		return
	}

	username := os.Args[3]

	if _, exists := auth.Users[username]; !exists {
		fmt.Printf("user %s does not exist\n", username)
		return
	}

	fmt.Print("new password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("error reading password: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Print("confirm password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("error reading password: %v\n", err)
		return
	}
	fmt.Println()

	if string(password) != string(confirmPassword) {
		fmt.Println("passwords do not match")
		return
	}

	err = auth.ChangePassword(username, string(password))
	if err != nil {
		fmt.Printf("error changing password: %v\n", err)
		return
	}

	fmt.Printf("password changed for user %s\n", username)
}

func handleAddRole(auth *authenticator.Authenticator) {
	if len(os.Args) < 5 {
		fmt.Println("usage: quiki auth add-role <username> <role>")
		return
	}

	username := os.Args[3]
	role := os.Args[4]

	if _, exists := auth.Users[username]; !exists {
		fmt.Printf("user %s does not exist\n", username)
		return
	}

	availableRoles := auth.GetAvailableRoles()
	if _, exists := availableRoles[role]; !exists {
		fmt.Printf("role %s does not exist\n", role)
		return
	}

	err := auth.AddUserRole(username, role)
	if err != nil {
		fmt.Printf("error adding role: %v\n", err)
		return
	}

	fmt.Printf("role %s added to user %s\n", role, username)
}

func handleRemoveRole(auth *authenticator.Authenticator) {
	if len(os.Args) < 5 {
		fmt.Println("usage: quiki auth remove-role <username> <role>")
		return
	}

	username := os.Args[3]
	role := os.Args[4]

	if _, exists := auth.Users[username]; !exists {
		fmt.Printf("user %s does not exist\n", username)
		return
	}

	err := auth.RemoveUserRole(username, role)
	if err != nil {
		fmt.Printf("error removing role: %v\n", err)
		return
	}

	fmt.Printf("role %s removed from user %s\n", role, username)
}

func handleListRoles(auth *authenticator.Authenticator) {
	availableRoles := auth.GetAvailableRoles()
	if len(availableRoles) == 0 {
		fmt.Println("no roles found")
		return
	}

	fmt.Println("roles:")
	for roleName, role := range availableRoles {
		fmt.Printf("  %s", roleName)
		if len(role.Permissions) > 0 {
			fmt.Printf(" (permissions: %v)", role.Permissions)
		}
		if len(role.Inherits) > 0 {
			fmt.Printf(" (inherits: %v)", role.Inherits)
		}
		fmt.Println()
	}
}

func handleMapUser(auth *authenticator.Authenticator) {
	if !auth.IsServer {
		fmt.Println("map-user command only available for server auth (don't use -wiki flag)")
		return
	}

	if len(os.Args) < 6 {
		fmt.Println("usage: quiki auth map-user <server-user> <wiki-name> <wiki-username>")
		return
	}

	serverUser := os.Args[3]
	wikiName := os.Args[4]
	wikiUsername := os.Args[5]

	if _, exists := auth.Users[serverUser]; !exists {
		fmt.Printf("server user %s does not exist\n", serverUser)
		return
	}

	err := auth.MapUser(serverUser, wikiName, wikiUsername)
	if err != nil {
		fmt.Printf("error mapping user: %v\n", err)
		return
	}

	fmt.Printf("mapped server user %s to wiki user %s in wiki %s\n", serverUser, wikiUsername, wikiName)
}

func handleUnmapUser(auth *authenticator.Authenticator) {
	if !auth.IsServer {
		fmt.Println("unmap-user command only available for server auth (don't use -wiki flag)")
		return
	}

	if len(os.Args) < 5 {
		fmt.Println("usage: quiki auth unmap-user <server-user> <wiki-name>")
		return
	}

	serverUser := os.Args[3]
	wikiName := os.Args[4]

	err := auth.UnmapUser(serverUser, wikiName)
	if err != nil {
		fmt.Printf("error unmapping user: %v\n", err)
		return
	}

	fmt.Printf("unmapped server user %s from wiki %s\n", serverUser, wikiName)
}

func handleListMappings(auth *authenticator.Authenticator) {
	if !auth.IsServer {
		fmt.Println("list-mappings command only available for server auth (don't use -wiki)")
		return
	}

	// if specific user provided, show only their mappings
	if len(os.Args) >= 4 {
		serverUser := os.Args[3]
		wikis := auth.GetUserWikis(serverUser)

		if len(wikis) == 0 {
			fmt.Printf("no wiki mappings found for user %s\n", serverUser)
			return
		}

		fmt.Printf("wiki mappings for user %s:\n", serverUser)
		for _, wiki := range wikis {
			wikiUsername, _ := auth.GetWikiUsername(serverUser, wiki)
			fmt.Printf("  %s -> %s\n", wiki, wikiUsername)
		}
		return
	}

	// show all mappings for all users
	fmt.Println("all wiki mappings:")
	hasAnyMappings := false

	for username, user := range auth.Users {
		if len(user.Wikis) > 0 {
			hasAnyMappings = true
			for wikiName, wikiUsername := range user.Wikis {
				fmt.Printf("  %s:%s -> %s\n", username, wikiName, wikiUsername)
			}
		}
	}

	if !hasAnyMappings {
		fmt.Println("no wiki mappings found")
	}
}
