package handler

import (
	"context"
	"encoding/json"
	
	"net/http"
	"os"
	"sync"
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// In-memory token store (replace with DB in production)
var (
	tokenStore = make(map[string]*oauth2.Token)
	storeMutex sync.Mutex

	sessionStore = make(map[string]string) // sessionToken -> githubAccessToken
	sessionMutex sync.Mutex
)

var githubOauthConfig *oauth2.Config

func InitOAuthConfig() {
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"repo", "user"},
		Endpoint:     github.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
	}
}

// Helper to generate a random session token
func generateSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Exported function to access sessionStore
func SessionStore() map[string]string {
	return sessionStore
}

// HandleGitHubLogin redirects user to GitHub for login
func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := "random" // TODO: generate and validate state for CSRF protection
	url := githubOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleGitHubCallback handles GitHub's redirect with ?code=...
func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID := token.AccessToken

	storeMutex.Lock()
	tokenStore[userID] = token
	storeMutex.Unlock()

	// Generate a session token and store mapping
	sessionToken := generateSessionToken()
	sessionMutex.Lock()
	sessionStore[sessionToken] = userID
	sessionMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_token": sessionToken,
	})
}

// GetTokenForUser retrieves a token for a user (replace with DB lookup)
func GetTokenForUser(userID string) *oauth2.Token {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	return tokenStore[userID]
} 