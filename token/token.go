package token

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	JwtSecret []byte
}

func (t *Token) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.JwtSecret)
}

func (t *Token) ParseToken(tokenString string) (int, error) {
	parsedToken, err := jwt.Parse(tokenString, func(tok *jwt.Token) (interface{}, error) {
		if _, err := tok.Method.(*jwt.SigningMethodHMAC); !err {
			return nil, fmt.Errorf("Unexpected signing method: %v", tok.Header["alg"])
		}
		return t.JwtSecret, nil
	})
	if err != nil {
		return 0, nil
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user_id claim")
	}
	return int(userIDFloat), nil
}
