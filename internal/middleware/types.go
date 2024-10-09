package middleware

import "github.com/golang-jwt/jwt"

type Claims struct {
	Role      uint8  `json:"role"`
	Id        uint64 `json:"id"`
	UserAgent string `json:"userAgent"`
	jwt.StandardClaims
}

type RefreshClaims struct {
	Role      uint8
	Id        uint64
	UserAgent string
	jwt.StandardClaims
}
