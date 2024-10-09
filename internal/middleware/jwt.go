package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
	SecretKey       = []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE")
)

type JWTHandler struct {
	// access token key
	AtKey []byte
	// refresh token key
	RtKey []byte
}

func NewJWTHandler() *JWTHandler {
	return &JWTHandler{
		AtKey: SecretKey,
		RtKey: SecretKey,
	}
}

func (js *JWTHandler) GenerateToken(ctx *gin.Context, role uint8, id uint64, userAgent string) error {
	claims := Claims{
		Role: role,
		Id:   id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			Issuer:    "oj",
		},
		UserAgent: userAgent,
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	ctx.Header("x-jwt-token", token)
	return err
}

func (js *JWTHandler) RefreshToken(ctx *gin.Context, role uint8, id uint64, userAgent string) error {
	claims := RefreshClaims{
		Role: role,
		Id:   id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "oj",
		},
		UserAgent: userAgent,
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	ctx.Header("x-refresh-token", token)
	return err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenInvalid
			} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&(jwt.ValidationErrorNotValidYet) != 0 {
				return nil, ErrLoginYet
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
	case errors.Is(err, ErrLoginYet):
		code = http.StatusUnauthorized
		msg = "have not logged in yet"
	default:
		code = http.StatusInternalServerError
		msg = "parse Token failed"
	}
	return code, msg
}
