package jwt

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	SecretKey = []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE")
	// AtKey access token key
	AtKey = SecretKey
	// RtKey refresh token key
	RtKey = SecretKey
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) *RedisJWTHandler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, role uint8, uid uint64) error {
	ssid := uuid.New().String()
	err := h.AccessToken(ctx, role, uid, ssid)
	if err != nil {
		return err
	}

	err = h.RefreshToken(ctx, role, uid, ssid)
	if err != nil {
		return err
	}

	return nil
}

func (h *RedisJWTHandler) AccessToken(ctx *gin.Context, role uint8, id uint64, ssid string) error {
	claims := Claims{
		Role: role,
		Id:   id,
		SSId: ssid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			Issuer:    "oj",
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	ctx.Header("x-jwt-token", token)
	return err
}

func (h *RedisJWTHandler) RefreshToken(ctx *gin.Context, role uint8, id uint64, ssid string) error {
	claims := RefreshClaims{
		Role: role,
		Id:   id,
		SSId: ssid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "oj",
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	ctx.Header("x-refresh-token", token)
	return err
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	// 检查请求头中是否包含 Token
	if tokenHeader == "" {
		return ""
	}
	return tokenHeader
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	_, err := h.cmd.Exists(ctx.Request.Context(), fmt.Sprintf("user:ssid:%s", ssid)).Result()
	return err
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	claim := ctx.MustGet("claims").(*Claims)
	return h.cmd.Set(ctx.Request.Context(), fmt.Sprintf("user:ssid:%s", claim.SSId), "", time.Hour*24*7).Err()
}
