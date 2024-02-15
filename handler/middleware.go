package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gearyaudie/mpx-backend.git/utils"
)

func RequireLogin(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the token is in the format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate the token
		token := tokenParts[1]
		claims, err := utils.VerifyToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Proceed with the handler function if authentication is successful
		handlerFunc(w, r.WithContext(context.WithValue(r.Context(), "user", claims.UserID)))
	}
}
