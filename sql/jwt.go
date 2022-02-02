package sql

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

var JwtSigningKey = []byte("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")

func GetJWTFromUserPass(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"iss":   "MeetPlanCA",
		"exp":   expirationTime.Unix(),
	})

	return token.SignedString(JwtSigningKey)
}

func CheckJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return JwtSigningKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] == "MeetPlanCA" {
			return claims, nil
		}
		return nil, errors.New("JWT issuer isn't correct")
	} else {
		return nil, err
	}
}
