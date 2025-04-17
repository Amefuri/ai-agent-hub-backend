package middleware

import (
	"os"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// func GenerateJWT(secretKey string, userID string) (string, error) {
// 	claims := jwt.MapClaims{
// 		"sub": userID,
// 		"iat": time.Now().Unix(),
// 		"exp": time.Now().Add(time.Hour * 24).Unix(),
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString([]byte(secretKey))
// }

// func ValidateJWT(tokenString string, secretKey string) (*jwt.Token, error) {
// 	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(secretKey), nil
// 	})
// }

func JWTMiddleware() echo.MiddlewareFunc {
    secret := os.Getenv("JWT_SECRET")
    return echojwt.WithConfig(echojwt.Config{
        SigningKey: []byte(secret),
    })
}

