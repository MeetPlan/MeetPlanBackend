package sql

import (
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/golang-jwt/jwt"
	"strconv"
	"strings"
	"time"
)

func GetSigningKey() []byte {
	config, err := GetConfig()
	if err != nil || !config.Debug {
		return []byte(uniuri.NewLen(100))
	}
	return []byte("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")
}

var JwtSigningKey = GetSigningKey()

const JWTIssuer = "MeetPlanCA"

func GetJWTFromUserPass(email string, role string, uid int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uid,
		"email":   email,
		"role":    role,
		"iss":     JWTIssuer,
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
		"iss":     JWTIssuer,
		"exp":     expirationTime.Unix(),
	})

	sgnd, err := token.SignedString(JwtSigningKey)
	return sgnd, err, strings.Split(expirationTime.Format("02-01-2006"), " ")[0]
}

func (db *sqlImpl) CheckJWT(tokenString string) (user User, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return JwtSigningKey, nil
	})

	if token == nil {
		return user, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] != JWTIssuer {
			return user, errors.New("JWT issuer isn't correct")
		}
		userId, err := strconv.Atoi(fmt.Sprint(claims["user_id"]))
		if err != nil {
			return user, err
		}
		user, err = db.GetUser(userId)
		return user, err
	} else {
		return user, err
	}
}
