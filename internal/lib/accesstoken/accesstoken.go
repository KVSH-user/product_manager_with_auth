package accesstoken

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func Generate(secretKey string, uid int, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = uid
	claims["exp"] = time.Now().Add(ttl).Unix()
	claims["issued"] = time.Now().Unix()

	return token.SignedString([]byte(secretKey))
}
