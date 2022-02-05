package sql

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"strings"
	"time"
)

var JwtSigningKey = []byte("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")

func GetJWTFromUserPass(email string, role string, uid int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uid,
		"email":   email,
		"role":    role,
		"iss":     "MeetPlanCA",
		"exp":     expirationTime.Unix(),
	})

	return token.SignedString(JwtSigningKey)
}

func GetJWTForTestingResult(userId int, result string, testId int, date string) (string, error, string) {
	expirationTime, err := time.Parse("02-01-2006", date)
	if err != nil {
		return "", err, ""
	}
	expirationTime = expirationTime.Add(48 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"result":  result,
		"test_id": testId,
		"iss":     "MeetPlanCA",
		"exp":     expirationTime.Unix(),
	})

	sgnd, err := token.SignedString(JwtSigningKey)
	return sgnd, err, strings.Split(expirationTime.Format("02-01-2006"), " ")[0]
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
