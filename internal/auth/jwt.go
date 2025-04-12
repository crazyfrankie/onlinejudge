package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
)

type JWTHandler interface {
	SetLoginToken(ctx *gin.Context, uid uint64) error
	AccessToken(ctx *gin.Context, id uint64, ssid string) (string, error)
	RefreshToken(ctx *gin.Context, id uint64, ssid string) (string, error)
	ExtractToken(ctx *gin.Context) string
	ExtractRefreshToken(ctx *gin.Context) string
	ParseToken(token string) (*Claims, error)
	CheckSession(ctx *gin.Context, uid uint64, ssid string) error
	ClearToken(ctx *gin.Context) error
	TryRefresh(ctx *gin.Context) error
	LogoutAllDevices(ctx *gin.Context) error
	HandleTokenError(err error) *er.BizError
}

type Claims struct {
	Id        uint64 `json:"id"`
	UserAgent string `json:"userAgent"`
	SSId      string
	jwt.StandardClaims
}

type RefreshClaims struct {
	Role      uint8
	Id        uint64
	UserAgent string
	SSId      string
	jwt.StandardClaims
}

var (
	SecretKey = []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE")
	// AtKey access token key
	AtKey = SecretKey
	// RtKey refresh token key
	RtKey = SecretKey

	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
	ErrLoginYet     = errors.New("have not logged in yet")
)

const (
	ssKey = "user:%d:ssid:%s:%s"
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) JWTHandler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid uint64) error {
	ssid := uuid.New().String()
	accessToken, err := h.AccessToken(ctx, uid, ssid)
	if err != nil {
		return err
	}

	refreshToken, err := h.RefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}

	ctx.SetCookie("access_token", accessToken, 900, "/", "", false, false)
	ctx.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true)

	ua := ctx.GetHeader("User-Agent")
	device := hashUA(ua)
	key := fmt.Sprintf(ssKey, uid, ssid, device)

	err = h.cmd.Set(ctx.Request.Context(),
		key, refreshToken, time.Hour*24*7).Err()
	if err != nil {
		return err
	}

	// 把设备记录入集合
	deviceSetKey := fmt.Sprintf("user:%d:devices", uid)
	session := fmt.Sprintf("%s:%s", ssid, device)

	return h.cmd.SAdd(ctx.Request.Context(), deviceSetKey, session).Err()
}

func (h *RedisJWTHandler) AccessToken(ctx *gin.Context, id uint64, ssid string) (string, error) {
	claims := Claims{
		Id:   id,
		SSId: ssid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			Issuer:    "github.com/crazyfrankie/onlinejudge",
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	return token, err
}

func (h *RedisJWTHandler) RefreshToken(ctx *gin.Context, id uint64, ssid string) (string, error) {
	claims := RefreshClaims{
		Id:   id,
		SSId: ssid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "github.com/crazyfrankie/onlinejudge",
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(SecretKey)
	return token, err
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	token, err := ctx.Cookie("access_token")
	if err != nil {
		return ""
	}

	return token
}

func (h *RedisJWTHandler) ExtractRefreshToken(ctx *gin.Context) string {
	token, err := ctx.Cookie("refresh_token")
	if err != nil {
		return ""
	}

	return token
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, uid uint64, ssid string) error {
	device := hashUA(ctx.GetHeader("User-Agent"))
	key := fmt.Sprintf(ssKey, uid, ssid, device)

	res, err := h.cmd.Get(ctx.Request.Context(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return er.NewBizError(constant.ErrUserInvalidToken)
		}
		return err
	}

	_, err = h.ParseToken(res)
	if err != nil {
		return h.HandleTokenError(err)
	}

	return nil
}

func (h *RedisJWTHandler) TryRefresh(ctx *gin.Context) error {
	refreshToken := h.ExtractRefreshToken(ctx)
	if refreshToken == "" {
		return nil
	}

	claims, err := h.ParseToken(refreshToken)
	if err != nil {
		return h.HandleTokenError(err)
	}

	device := hashUA(ctx.GetHeader("User-Agent"))
	key := fmt.Sprintf(ssKey, claims.Id, claims.SSId, device)

	storedToken, err := h.cmd.Get(ctx.Request.Context(), key).Result()
	if err != nil {
		return err
	}

	if storedToken != refreshToken {
		return er.NewBizError(constant.ErrUserInvalidToken)
	}

	ttl, err := h.cmd.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	refreshWindow := (7 * 24 * time.Hour) / 3
	minTTL := 5 * time.Minute

	if ttl < refreshWindow && ttl > minTTL {
		return h.SetLoginToken(ctx, claims.Id)
	}

	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.SetCookie("access_token", "", 900, "/", "", false, false)
	ctx.SetCookie("refresh_token", "", 7*24*3600, "/", "", false, true)

	claim := ctx.MustGet("claims").(*Claims)

	device := hashUA(ctx.GetHeader("User-Agent"))
	key := fmt.Sprintf(ssKey, claim.Id, claim.SSId, device)

	session := fmt.Sprintf("%s:%s", claim.SSId, device)
	deviceSetKey := fmt.Sprintf("user:%d:devices", claim.Id)

	pipe := h.cmd.TxPipeline()
	pipe.Del(ctx.Request.Context(), key)
	pipe.SRem(ctx.Request.Context(), deviceSetKey, session)
	_, err := pipe.Exec(ctx.Request.Context())
	return err
}

func (h *RedisJWTHandler) LogoutAllDevices(ctx *gin.Context) error {
	claim := ctx.MustGet("claims").(*Claims)

	deviceSetKey := fmt.Sprintf("user:%d:devices", claim.Id)
	sessions, err := h.cmd.SMembers(ctx.Request.Context(), deviceSetKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	pipe := h.cmd.TxPipeline()
	for _, session := range sessions {
		key := fmt.Sprintf("user:%d:ssid:%s", claim.Id, session)
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, deviceSetKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (h *RedisJWTHandler) ParseToken(token string) (*Claims, error) {
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

func (h *RedisJWTHandler) HandleTokenError(err error) *er.BizError {
	var errCode constant.ErrorCode

	switch {
	case errors.Is(err, ErrTokenInvalid):
		errCode = constant.ErrUserInvalidToken
	case errors.Is(err, ErrTokenExpired):
		errCode = constant.ErrUserTokenExpired
	case errors.Is(err, ErrLoginYet):
		errCode = constant.ErrUserLoginYet
	default:
		errCode = constant.ErrUserInternalServer
	}

	return er.NewBizError(errCode)
}

func hashUA(ua string) string {
	sum := sha1.Sum([]byte(ua))
	return hex.EncodeToString(sum[:])
}
