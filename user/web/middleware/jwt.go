package middleware

import (
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
)

type Claims struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	Authority  int    `json:"authority"`
	jwt.StandardClaims
}

func GenerateToken(identifier, password string, authority int) (string, error) {
	gob.Register(time.Now())
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)
	claims := Claims{
		Identifier: identifier,
		Password:   password,
		Authority:  authority,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "todolist",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE"))
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE"), nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenInvalid
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, ErrTokenExpired
			}
		}
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, ErrTokenInvalid
}

func handleTokenError(err error) (int, string) {
	var code int
	var msg string
	switch {
	case errors.Is(err, ErrTokenExpired):
		code = http.StatusUnauthorized
		msg = "token is expired"
	case errors.Is(err, ErrTokenInvalid):
		code = http.StatusUnauthorized
		msg = "token is invalid"
	default:
		code = http.StatusInternalServerError
		msg = "parse Token failed"
	}
	return code, msg
}
