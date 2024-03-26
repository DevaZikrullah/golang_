package middleware

import (
	"net/http"
	"test/utils"

	"github.com/dgrijalva/jwt-go"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil // replace "your-secret-key" with your actual secret key
		})

		if err != nil || !token.Valid {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
