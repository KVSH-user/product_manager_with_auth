package uidextractor

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

func ValidateToken(authHeader, secret string) (int, error) {
	mySigningKey := secret

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return 0, fmt.Errorf("invalid token format")
	}

	tokenString := splitToken[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(mySigningKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, ok := claims["uid"].(float64)
		if !ok {
			return 0, fmt.Errorf("uid claim is missing")
		}
		return int(uid), nil
	} else {
		return 0, fmt.Errorf("invalid token")
	}
}
