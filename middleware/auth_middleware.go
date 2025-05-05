package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"net/http"
	"strings"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"context"
)

func AuthMiddleware(next http.Handler) http.Handler {
	godotenv.Load()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.EnableCORS(&w)
		if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("jwtSecret")), nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			fmt.Println("Token parsing error:", err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["user_id"]
			if !ok {
				http.Error(w, "Invalid user ID in token", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	})
}