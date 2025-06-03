package middleware

import (
	"net/http"
	"strings"
	"tersoh-backend/internal/utils"
	"tersoh-backend/models"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("your-256-bit-secret")

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			utils.RespondError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}
		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
