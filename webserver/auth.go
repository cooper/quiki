package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cooper/quiki/authenticator"
)

// rateLimitKey represents a composite key for rate limiting
type rateLimitKey struct {
	ip        string
	username  string
	userAgent string
}

// attemptInfo tracks failed attempts
type attemptInfo struct {
	count      int
	lastTry    time.Time
	blockUntil time.Time
	userAgent  string
	username   string
}

var (
	// ip-based limits (loose)
	ipAttempts = make(map[string]*attemptInfo)

	// composite limits (moderate - prevent coordinated attacks)
	compositeAttempts = make(map[rateLimitKey]*attemptInfo)
	rateLimitMux      sync.RWMutex

	// username-based limits (strict - prevent account targeting)
	usernameAttempts = make(map[string]*attemptInfo)
)

// SetSecurityHeaders adds security headers to prevent common attacks
func SetSecurityHeaders(w http.ResponseWriter) {
	// prevent clickjacking
	w.Header().Set("X-Frame-Options", "DENY")

	// prevent mime type sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// referrer policy for privacy
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// strict transport security for https enforcement (1 year, include subdomains)
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	// prevent cross-domain policy abuse
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

	// content security policy - tightened for production security
	// removed unsafe-inline for scripts to prevent xss attacks
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; object-src 'none'; base-uri 'self'; form-action 'self'")
}

// SanitizeInput sanitizes user input to prevent xss attacks
func SanitizeInput(input string) string {
	return html.EscapeString(strings.TrimSpace(input))
}

func generateCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("csrf_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// ValidateCSRFToken checks if the submitted csrf token matches the session token
func ValidateCSRFToken(r *http.Request, sessionToken string) bool {
	return validateCSRFToken(r, sessionToken)
}

func validateCSRFToken(r *http.Request, sessionToken string) bool {
	formToken := r.FormValue("csrf_token")
	return formToken != "" && formToken == sessionToken
}

// GetOrCreateCSRFToken gets existing csrf token or creates new one for session
func GetOrCreateCSRFToken(r *http.Request) string {
	return getOrCreateCSRFToken(r)
}

func getOrCreateCSRFToken(r *http.Request) string {
	if token := SessMgr.GetString(r.Context(), "csrf_token"); token != "" {
		return token
	}
	token := generateCSRFToken()
	SessMgr.Put(r.Context(), "csrf_token", token)
	return token
}

// create fingerprint to represent the request, used to detect anomalies
func extractRequestFingerprint(r *http.Request) string {
	var parts []string

	// user agent (most stable)
	if ua := r.Header.Get("User-Agent"); ua != "" {
		parts = append(parts, ua)
	}

	// accept languages (fairly stable)
	if al := r.Header.Get("Accept-Language"); al != "" {
		parts = append(parts, al)
	}

	// accept encoding (fairly stable)
	if ae := r.Header.Get("Accept-Encoding"); ae != "" {
		parts = append(parts, ae)
	}

	// create hash of combined characteristics
	fingerprint := strings.Join(parts, "|")
	bytes := make([]byte, 16)
	copy(bytes, []byte(fingerprint))
	return hex.EncodeToString(bytes)[:16] // truncate to reasonable length
}

// detectSuspiciousActivity analyzes request patterns for anomalies
func detectSuspiciousActivity(r *http.Request, username string) bool {
	fingerprint := extractRequestFingerprint(r)
	ip := GetClientIP(r)

	rateLimitMux.RLock()
	defer rateLimitMux.RUnlock()

	// check if this username has been attempted from many different fingerprints recently
	suspiciousCount := 0
	for key, attempt := range compositeAttempts {
		if key.username == strings.ToLower(username) &&
			time.Since(attempt.lastTry) < 10*time.Minute {

			// different fingerprint but same username = p suspicious
			if extractRequestFingerprint(&http.Request{
				Header: http.Header{"User-Agent": {attempt.userAgent}},
			}) != fingerprint {
				suspiciousCount++
			}
		}
	}

	// if username attempted from 3+ different fingerprints in 10 minutes, also somewhat suspicious
	if suspiciousCount >= 3 {
		log.Printf("suspicious activity detected: username %s attempted from %d different fingerprints from ip %s",
			username, suspiciousCount+1, ip)
		return true
	}

	return false
}
func GetClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = strings.Split(r.RemoteAddr, ":")[0]
	}
	return clientIP
}

// ValidateAuthForm validates authentication form data
func ValidateAuthForm(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}
	return nil
}

// CheckRateLimit checks if a request should be rate limited
func CheckRateLimit(r *http.Request, username string) bool {
	return checkRateLimit(r, username)
}

func addRateLimitHeaders(w http.ResponseWriter, r *http.Request, username string) {
	rateLimitMux.RLock()
	defer rateLimitMux.RUnlock()

	ip := GetClientIP(r)
	username = strings.ToLower(username)
	compositeKey := extractRateLimitKey(r, username)

	// get current attempt counts
	var ipCount, userCount, compCount int
	var ipRemaining, userRemaining, compRemaining int

	if ipAttempt, exists := ipAttempts[ip]; exists {
		ipCount = ipAttempt.count
	}
	if userAttempt, exists := usernameAttempts[username]; exists {
		userCount = userAttempt.count
	}
	if compAttempt, exists := compositeAttempts[compositeKey]; exists {
		compCount = compAttempt.count
	}

	// calculate remaining attempts before blocking
	ipRemaining = max(0, 20-ipCount)
	userRemaining = max(0, 5-userCount)
	compRemaining = max(0, 10-compCount)

	// set headers with most restrictive limits
	minRemaining := min(min(ipRemaining, userRemaining), compRemaining)
	w.Header().Set("X-RateLimit-Limit", "5") // most restrictive limit (username-based)
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", minRemaining))

	if minRemaining == 0 {
		w.Header().Set("Retry-After", "300") // 5 minutes in seconds
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// HandleAuthError handles authentication errors with rate limiting
func HandleAuthError(w http.ResponseWriter, err error, r *http.Request, username string) {
	recordLoginFail(r, username)

	// add rate limiting info headers
	addRateLimitHeaders(w, r, username)

	http.Error(w, "invalid username or password", http.StatusUnauthorized)
}

// ClearSuccessfulLogin clears failed attempts after successful login
func ClearSuccessfulLogin(r *http.Request, username string) {
	ip := GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// log successful login for security monitoring
	log.Printf("auth: successful login for user=%s from ip=%s ua=%s",
		username, ip, userAgent[:min(len(userAgent), 50)])

	clearSuccessfulLogin(r, username)
}

func extractRateLimitKey(r *http.Request, username string) rateLimitKey {
	return rateLimitKey{
		ip:        GetClientIP(r),
		username:  strings.ToLower(username),
		userAgent: r.Header.Get("User-Agent"),
	}
}

// checkRateLimit implements multi-factor rate limiting
// returns true if request should be blocked
func checkRateLimit(r *http.Request, username string) bool {
	rateLimitMux.RLock()
	defer rateLimitMux.RUnlock()

	now := time.Now()
	ip := GetClientIP(r)
	username = strings.ToLower(username)
	compositeKey := extractRateLimitKey(r, username)

	// check for suspicious activity patterns first
	if detectSuspiciousActivity(r, username) {
		return true
	}

	// check ip-based limiting (very lenient, gotta consider shared networks)
	if ipAttempt, exists := ipAttempts[ip]; exists && now.Before(ipAttempt.blockUntil) {
		// only block if there are excessive attempts from this ip
		if ipAttempt.count > 20 { // high threshold for shared networks
			return true
		}
	}

	// check username-based limiting (strict - prevents account targeting)
	if userAttempt, exists := usernameAttempts[username]; exists && now.Before(userAttempt.blockUntil) && userAttempt.count > 5 {
		return true
	}

	// check composite limiting (moderate - prevents coordinated attacks)
	if compAttempt, exists := compositeAttempts[compositeKey]; exists && now.Before(compAttempt.blockUntil) && compAttempt.count > 10 {
		return true
	}

	return false
}

func recordLoginFail(r *http.Request, username string) {
	rateLimitMux.Lock()
	defer rateLimitMux.Unlock()

	now := time.Now()
	ip := GetClientIP(r)
	username = strings.ToLower(username)
	userAgent := r.Header.Get("User-Agent")
	compositeKey := extractRateLimitKey(r, username)

	log.Printf("auth: failed login attempt for user=%s from ip=%s ua=%s",
		username, ip, userAgent[:min(len(userAgent), 50)])

	// record ip-based attempt
	ipAttempt := ipAttempts[ip]
	if ipAttempt == nil {
		ipAttempt = &attemptInfo{}
		ipAttempts[ip] = ipAttempt
	}
	ipAttempt.count++
	ipAttempt.lastTry = now
	ipAttempt.userAgent = userAgent

	// fairly lenient backoff for ip
	var ipBlockDuration time.Duration
	switch {
	case ipAttempt.count <= 10:
		ipBlockDuration = 5 * time.Minute
	case ipAttempt.count <= 20:
		ipBlockDuration = 15 * time.Minute
	default:
		ipBlockDuration = 1 * time.Hour // only block after many attempts
		log.Printf("auth: ip %s blocked after %d failed attempts", ip, ipAttempt.count)
	}
	ipAttempt.blockUntil = now.Add(ipBlockDuration)

	// record username-based attempt
	userAttempt := usernameAttempts[username]
	if userAttempt == nil {
		userAttempt = &attemptInfo{}
		usernameAttempts[username] = userAttempt
	}
	userAttempt.count++
	userAttempt.lastTry = now
	userAttempt.username = username

	// more strict backoff for username targeting
	var userBlockDuration time.Duration
	switch {
	case userAttempt.count <= 3:
		userBlockDuration = 1 * time.Minute
	case userAttempt.count <= 5:
		userBlockDuration = 5 * time.Minute
		log.Printf("auth: user %s temporarily blocked after %d failed attempts", username, userAttempt.count)
	case userAttempt.count <= 8:
		userBlockDuration = 15 * time.Minute
		log.Printf("auth: user %s blocked for 15min after %d failed attempts", username, userAttempt.count)
	default:
		userBlockDuration = 2 * time.Hour // harsher penalty for account targeting, f off
		log.Printf("auth: user %s blocked for 2h after %d failed attempts - possible brute force", username, userAttempt.count)
	}
	userAttempt.blockUntil = now.Add(userBlockDuration)

	// record composite attempt
	compAttempt := compositeAttempts[compositeKey]
	if compAttempt == nil {
		compAttempt = &attemptInfo{}
		compositeAttempts[compositeKey] = compAttempt
	}
	compAttempt.count++
	compAttempt.lastTry = now
	compAttempt.userAgent = userAgent
	compAttempt.username = username

	// moderate backoff for other request characteristics
	var compBlockDuration time.Duration
	switch {
	case compAttempt.count <= 5:
		compBlockDuration = 2 * time.Minute
	case compAttempt.count <= 10:
		compBlockDuration = 10 * time.Minute
		log.Printf("auth: composite key (ip=%s, user=%s) blocked after %d attempts", ip, username, compAttempt.count)
	case compAttempt.count <= 15:
		compBlockDuration = 30 * time.Minute
		log.Printf("auth: coordinated attack detected for ip=%s, user=%s - 30min block", ip, username)
	default:
		compBlockDuration = 90 * time.Minute // significant penalty for persistent attacks
		log.Printf("auth: persistent attack from ip=%s targeting user=%s - 90min block", ip, username)
	}
	compAttempt.blockUntil = now.Add(compBlockDuration)

	// cleanup old entries periodically
	cleanupOldRateLimitEntries()
}

// clearSuccessfulLogin clears attempts after successful login
func clearSuccessfulLogin(r *http.Request, username string) {
	rateLimitMux.Lock()
	defer rateLimitMux.Unlock()

	ip := GetClientIP(r)
	username = strings.ToLower(username)
	compositeKey := extractRateLimitKey(r, username)

	// clear or reduce attempts rather than completely removing
	// this maintains some memory of previous suspicious activity

	if ipAttempt, exists := ipAttempts[ip]; exists {
		if ipAttempt.count <= 3 {
			delete(ipAttempts, ip) // clear minor attempts
		} else {
			ipAttempt.count = ipAttempt.count / 2 // reduce but remember persistent attempts
			ipAttempt.blockUntil = time.Time{}
		}
	}

	if userAttempt, exists := usernameAttempts[username]; exists {
		if userAttempt.count <= 3 {
			delete(usernameAttempts, username)
		} else {
			userAttempt.count = userAttempt.count / 2
			userAttempt.blockUntil = time.Time{}
		}
	}

	if compAttempt, exists := compositeAttempts[compositeKey]; exists {
		if compAttempt.count <= 3 {
			delete(compositeAttempts, compositeKey)
		} else {
			compAttempt.count = compAttempt.count / 2
			compAttempt.blockUntil = time.Time{}
		}
	}
}

// cleanupOldRateLimitEntries removes stale entries to prevent memory leaks
func cleanupOldRateLimitEntries() {
	cutoff := time.Now().Add(-24 * time.Hour)

	// only cleanup if maps are getting large
	if len(ipAttempts) > 1000 {
		for ip, attempt := range ipAttempts {
			if attempt.lastTry.Before(cutoff) && time.Now().After(attempt.blockUntil) {
				delete(ipAttempts, ip)
			}
		}
	}

	if len(usernameAttempts) > 500 {
		for username, attempt := range usernameAttempts {
			if attempt.lastTry.Before(cutoff) && time.Now().After(attempt.blockUntil) {
				delete(usernameAttempts, username)
			}
		}
	}

	if len(compositeAttempts) > 2000 {
		for key, attempt := range compositeAttempts {
			if attempt.lastTry.Before(cutoff) && time.Now().After(attempt.blockUntil) {
				delete(compositeAttempts, key)
			}
		}
	}
}

// authTemplateData provides common template data for auth pages
type authTemplateData struct {
	WikiName       string
	WikiLogo       string
	Static         string
	SharedStatic   string
	PageTitle      string
	Heading        string
	Error          string
	Success        string
	Username       string
	Email          string
	LoginAction    string
	RegisterAction string
	RegisterURL    string
	LoginURL       string
	HomeURL        string
	AllowRegister  bool
	ShowLinks      bool
	CSRFToken      string
}

// newAuthTemplateData creates base template data for auth pages
func (wi *WikiInfo) newAuthTemplateData(pageTitle, heading string, r *http.Request) authTemplateData {
	return authTemplateData{
		WikiName:      wi.Name,
		WikiLogo:      wi.Logo,
		Static:        wi.template.staticRoot,
		SharedStatic:  "/shared",
		PageTitle:     pageTitle,
		Heading:       heading,
		RegisterURL:   wi.Opt.Root.Wiki + "register",
		LoginURL:      wi.Opt.Root.Wiki + "login",
		HomeURL:       wi.Opt.Root.Wiki,
		AllowRegister: wi.Opt.Auth.Register,
		ShowLinks:     true,
		CSRFToken:     getOrCreateCSRFToken(r),
	}
}

// getWikiInfo extracts the wiki info from the request, using the same logic as handleRoot
func getWikiInfo(r *http.Request) *WikiInfo {
	var delayedWiki *WikiInfo

	// try each wiki (same logic as handleRoot)
	for _, w := range Wikis {
		// wrong root
		wikiRoot := w.Opt.Root.Wiki
		if r.URL.Path != wikiRoot && !strings.HasPrefix(r.URL.Path, wikiRoot+"/") {
			continue
		}

		// wrong host
		if w.Host != r.Host {
			// if the wiki host is empty, it is the fallback wiki
			if w.Host == "" && delayedWiki == nil {
				delayedWiki = w
			}
			continue
		}

		// host matches
		delayedWiki = w
		break
	}

	return delayedWiki
}

// requireAuth checks if the wiki requires authentication and redirects to login if needed
func requireAuth(wi *WikiInfo, w http.ResponseWriter, r *http.Request) bool {
	if wi.Opt.Auth.Require && !SessMgr.GetBool(r.Context(), "loggedIn") {
		// redirect to login with current page as redirect target
		loginURL := wi.Opt.Root.Wiki + "login?redirect=" + r.URL.Path
		http.Redirect(w, r, loginURL, http.StatusFound)
		return false // auth required, redirected
	}
	return true // auth not required or user is logged in
}

// handleLogin displays the login form or processes login attempts
func handleLogin(w http.ResponseWriter, r *http.Request) {
	wi := getWikiInfo(r)
	if wi == nil {
		http.NotFound(w, r)
		return
	}

	// if auth is not enabled, redirect to home
	if !wi.Opt.Auth.Enable {
		http.Redirect(w, r, wi.Opt.Root.Wiki, http.StatusFound)
		return
	}

	// get redirect target
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = wi.Opt.Root.Wiki
	}

	// if already logged in, redirect
	if SessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	// handle POST - login attempt
	if r.Method == "POST" {
		username := SanitizeInput(r.FormValue("username"))
		password := r.FormValue("password") // don't sanitize passwords

		if username == "" || password == "" {
			wi.showLoginForm(w, r, "username and password are required", redirect)
			return
		}

		// validate csrf token
		csrfToken := getOrCreateCSRFToken(r)
		if !validateCSRFToken(r, csrfToken) {
			wi.showLoginForm(w, r, "security token validation failed, please try again", redirect)
			return
		}

		// check rate limiting
		if checkRateLimit(r, username) {
			wi.showLoginForm(w, r, "too many failed attempts, please try again later", redirect)
			return
		}

		// authenticate against wiki's auth system
		wikiAuthPath := wi.Wiki.Dir("auth.json")
		wikiAuth, err := authenticator.Open(wikiAuthPath)
		if err != nil {
			log.Printf("login error: failed to open wiki auth: %v", err)
			wi.showLoginForm(w, r, "authentication system unavailable", redirect)
			return
		}

		user, err := wikiAuth.Login(username, password)
		if err != nil {
			recordLoginFail(r, username)
			wi.showLoginForm(w, r, "invalid username or password", redirect)
			return
		}

		// login successful - clear failed attempts and set session
		clearSuccessfulLogin(r, username)
		SessMgr.Put(r.Context(), "loggedIn", true)
		SessMgr.Put(r.Context(), "user", &user)
		SessMgr.Put(r.Context(), "wikiName", wi.Name)

		// regenerate session id after successful login for security
		if err := SessMgr.RenewToken(r.Context()); err != nil {
			log.Printf("failed to renew session token: %v", err)
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	// show login form
	wi.showLoginForm(w, r, "", redirect)
}

// handleLogout logs out the current user
func handleLogout(w http.ResponseWriter, r *http.Request) {
	wi := getWikiInfo(r)
	if wi == nil {
		http.NotFound(w, r)
		return
	}

	// clear session
	SessMgr.Remove(r.Context(), "loggedIn")
	SessMgr.Remove(r.Context(), "user")
	SessMgr.Remove(r.Context(), "wikiName")

	// redirect to home
	http.Redirect(w, r, wi.Opt.Root.Wiki, http.StatusFound)
}

// handleRegister displays the registration form or processes registration attempts
func handleRegister(w http.ResponseWriter, r *http.Request) {
	wi := getWikiInfo(r)
	if wi == nil {
		http.NotFound(w, r)
		return
	}

	// if auth or registration is not enabled, redirect to home
	if !wi.Opt.Auth.Enable || !wi.Opt.Auth.Register {
		http.NotFound(w, r)
		return
	}

	// if already logged in, redirect to home
	if SessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, wi.Opt.Root.Wiki, http.StatusFound)
		return
	}

	// handle POST - registration attempt
	if r.Method == "POST" {
		username := SanitizeInput(r.FormValue("username"))
		email := SanitizeInput(r.FormValue("email"))
		password := r.FormValue("password")                // don't sanitize passwords
		passwordConfirm := r.FormValue("password_confirm") // don't sanitize passwords

		// validation
		if username == "" || email == "" || password == "" {
			wi.showRegisterForm(w, r, "all fields are required", username, email)
			return
		}

		// validate csrf token
		csrfToken := getOrCreateCSRFToken(r)
		if !validateCSRFToken(r, csrfToken) {
			wi.showRegisterForm(w, r, "security token validation failed, please try again", username, email)
			return
		}

		if password != passwordConfirm {
			wi.showRegisterForm(w, r, "passwords do not match", username, email)
			return
		}

		if len(password) < 8 {
			wi.showRegisterForm(w, r, "password must be at least 8 characters", username, email)
			return
		}

		// create user in wiki's auth system
		wikiAuthPath := wi.Wiki.Dir("auth.json")
		wikiAuth, err := authenticator.Open(wikiAuthPath)
		if err != nil {
			log.Printf("register error: failed to open wiki auth: %v", err)
			wi.showRegisterForm(w, r, "authentication system unavailable", username, email)
			return
		}

		// check if user already exists
		if _, exists := wikiAuth.GetUser(username); exists {
			wi.showRegisterForm(w, r, "username already taken", "", email)
			return
		}

		// create the user with basic read permissions
		user, err := wikiAuth.CreateUser(username, password)
		if err != nil {
			log.Printf("register error: failed to create user: %v", err)
			wi.showRegisterForm(w, r, "failed to create account", username, email)
			return
		}

		// set email and permissions
		user.Email = email
		if len(user.Permissions) == 0 {
			user.Permissions = []string{"read.wiki"}
		}

		// use UpdateUser to persist changes to disk
		if err := wikiAuth.UpdateUser(username, user); err != nil {
			log.Printf("register error: failed to update user details: %v", err)
			wi.showRegisterForm(w, r, "failed to create account", username, email)
			return
		}

		// show success message
		wi.showRegisterSuccess(w, r)
		return
	}

	// show registration form
	wi.showRegisterForm(w, r, "", "", "")
}

// showLoginForm renders the login template
func (wi *WikiInfo) showLoginForm(w http.ResponseWriter, r *http.Request, errorMsg, redirect string) {
	data := wi.newAuthTemplateData("login", "welcome back", r)
	data.Error = errorMsg
	data.LoginAction = wi.Opt.Root.Wiki + "login?redirect=" + redirect

	if err := wi.template.template.ExecuteTemplate(w, "login.tpl", data); err != nil {
		log.Printf("failed to render login template: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// showRegisterForm renders the registration template
func (wi *WikiInfo) showRegisterForm(w http.ResponseWriter, r *http.Request, errorMsg, username, email string) {
	data := wi.newAuthTemplateData("register", "join "+wi.Name, r)
	data.Error = errorMsg
	data.Username = username
	data.Email = email
	data.RegisterAction = wi.Opt.Root.Wiki + "register"

	if err := wi.template.template.ExecuteTemplate(w, "register.tpl", data); err != nil {
		log.Printf("failed to render register template: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// showRegisterSuccess shows a success message after registration
func (wi *WikiInfo) showRegisterSuccess(w http.ResponseWriter, r *http.Request) {
	data := wi.newAuthTemplateData("register", "join "+wi.Name, r)
	data.Success = "account created successfully! you can now log in."
	data.RegisterAction = wi.Opt.Root.Wiki + "register"

	if err := wi.template.template.ExecuteTemplate(w, "register.tpl", data); err != nil {
		log.Printf("failed to render register success template: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
