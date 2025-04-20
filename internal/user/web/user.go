package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	"github.com/crazyfrankie/onlinejudge/internal/auth"
	"github.com/crazyfrankie/onlinejudge/internal/sm"
	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
)

type UserHandler struct {
	logger  *zapx.Logger
	userSvc service.UserService
	codeSvc sm.SmSvc
	jwt     auth.JWTHandler
}

func NewUserHandler(logger *zapx.Logger, userSvc service.UserService, codeSvc sm.SmSvc, jwtHdl auth.JWTHandler) *UserHandler {
	return &UserHandler{
		logger:  logger,
		userSvc: userSvc,
		codeSvc: codeSvc,
		jwt:     jwtHdl,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("api/user")
	{
		userGroup.POST("login", ctl.IdentifierLogin())
		userGroup.POST("logout", ctl.Logout())
		userGroup.POST("send-code", ctl.SendVerificationCode())
		userGroup.POST("verify-code", ctl.VerificationCode())
		userGroup.GET("info", ctl.GetUserInfo())
		userGroup.POST("refresh-token", ctl.TokenRefresh())
		userGroup.PATCH("update", ctl.UpdateInfo())
		userGroup.PATCH("update-pwd", ctl.UpdatePassword())
	}
}

func (ctl *UserHandler) SendVerificationCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendCodeReq
		if err := c.Bind(&req); err != nil {
			return
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrUserInvalidParams))
			return
		}

		err := ctl.codeSvc.Send(c.Request.Context(), req.Biz, req.Phone)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) VerificationCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := "onlinejudge/User/VerificationCode"
		var req VerifyCodeReq
		if err := c.Bind(&req); err != nil {
			ctl.logger.Error(c.Request.Context(), name, "bind req error", zap.Error(err))
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := ctl.codeSvc.Verify(ctx, req.Biz, req.Phone, req.Code)
		if err != nil {
			ctl.logger.Error(c.Request.Context(), name, "verify code error", zap.Error(err))
			response.Error(c, err)
			return
		}

		user, err := ctl.userSvc.FindOrCreateUser(c.Request.Context(), req.Phone)
		if err != nil {
			ctl.logger.Error(c.Request.Context(), name, "find or create user error", zap.Error(err))
			response.Error(c, err)
			return
		}

		err = ctl.jwt.SetLoginToken(c, user.Id)
		if err != nil {
			ctl.logger.Error(c.Request.Context(), name, "set login token error", zap.Error(err))
			response.Error(c, err)
			return
		}

		ctl.logger.Info(c.Request.Context(), name, fmt.Sprintf("%s:用户处理成功", req.Biz),
			zap.String("phone", req.Phone))

		response.Success(c, map[string]interface{}{
			"id":    user.Id,
			"phone": user.Phone,
			"name":  user.Name,
		})
	}
}

func (ctl *UserHandler) IdentifierLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginReq
		if err := c.Bind(&req); err != nil {
			return
		}

		validate := validator.New()
		isEmail := validate.Var(req.Identifier, "email") == nil

		user, err := ctl.userSvc.Login(c.Request.Context(), req.Identifier, req.Password, isEmail)
		if err != nil {
			response.Error(c, err)
			return
		}

		err = ctl.jwt.SetLoginToken(c, user.Id)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims")
		claim := claims.(*auth.Claims)

		user, err := ctl.userSvc.GetInfo(c.Request.Context(), claim.Id)
		if err != nil {
			response.Error(c, err)
		}

		response.Success(c, user)
	}
}

func (ctl *UserHandler) UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdatePwdReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims")
		claim := claims.(*auth.Claims)

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrUserInvalidParams))
			return
		}

		err := ctl.userSvc.UpdatePassword(c.Request.Context(), claim.Id, req.Password)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateInfoReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*auth.Claims)
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBizError(constant.ErrUserInvalidParams))
			return
		}
		bir, err := time.Parse(time.DateTime, req.Birthday)
		if err != nil {
			response.Error(c, errors.NewBizError(constant.ErrUserInvalidParams))
			return
		}
		err = ctl.userSvc.UpdateInfo(c.Request.Context(), domain.User{
			Id:       claims.Id,
			Name:     req.Name,
			Email:    req.Email,
			Birthday: bir,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

// TokenRefresh 可以同时刷新长短 toke，用 redis 来记录是否有，即 refresh_token 是一次性
// 参考登录部分，比较 User-Agent 来增强安全性
func (ctl *UserHandler) TokenRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只有这个接口拿出来的才是 refresh_token
		refreshToken := ctl.jwt.ExtractRefreshToken(c)
		var rc auth.RefreshClaims
		token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
			return auth.AtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = ctl.jwt.CheckSession(c, rc.Id, rc.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置新的 access_token
		_, err = ctl.jwt.AccessToken(c, rc.Id, rc.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ctl.jwt.ClearToken(c)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}
