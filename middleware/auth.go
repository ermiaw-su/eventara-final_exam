package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JwtSecret is the secret key used to sign and verify JWT tokens
var JwtSecret = []byte("SECRETKEY")

// contextKey is a type used to store user ID in the request context
type contextKey string

// UserIDKey is the key used to store user ID in the request context
const UserIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		authHeader := r.Header.Get("Authorization")

		// If the Authorization header is missing, return an error
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		// Split the token from the "Bearer " prefix
		splitToken := strings.Split(authHeader, "Bearer ")

		// If the token is not in the expected format, return an error
		if len(splitToken) != 2 {
			http.Error(w, `{"error": "Invalid token format"}`, http.StatusUnauthorized)
			return
		}

		// Take the second part of the token
		tokenString := splitToken[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})

		// If the token is invalid, return an error
		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Get the JWT
		claims := token.Claims.(jwt.MapClaims)

		// Get the user ID from the claims
		userID := int(claims["id"].(float64))

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// Pass the request to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}