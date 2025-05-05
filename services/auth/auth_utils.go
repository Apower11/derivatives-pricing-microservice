package auth

import(
	"github.com/golang-jwt/jwt/v5"
	"time"
	"os"
)

func generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), 
		"nbf":     time.Now().Unix(),                    
		"iat":     time.Now().Unix(),                    
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("jwtSecret")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}