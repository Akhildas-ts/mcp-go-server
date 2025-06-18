package middleware

import (
	"net/http"
	"strings"
	"mcpserver/internal/handler"
)

// AuthMiddleware checks for a valid session token in the Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		sessionStore := handler.SessionStore()
		if _, ok := sessionStore[sessionToken]; !ok {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
} 