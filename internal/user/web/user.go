package web

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	"github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/common/response"
	ijwt "github.com/crazyfrankie/onlinejudge/internal/auth/jwt"
	"github.com/crazyfrankie/onlinejudge/internal/user/service"
)

type UserHandler struct {
	userSvc service.UserService
	codeSvc service.CodeService
	ijwt.Handler
}

func NewUserHandler(userSvc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		codeSvc: codeSvc,
		Handler: jwtHdl,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("user")
	{
		userGroup.POST("login", ctl.IdentifierLogin())
		userGroup.POST("logout", ctl.LogOut())
		userGroup.POST("send-code", ctl.SendVerificationCode())
		userGroup.POST("verify-code", ctl.VerificationCode())
		userGroup.GET("info", ctl.GetUserInfo())
		userGroup.POST("refresh-token", ctl.TokenRefresh())
		userGroup.POST("name", ctl.UpdateName())
		userGroup.POST("email", ctl.UpdateEmail())
		userGroup.POST("password", ctl.UpdatePassword())
		userGroup.POST("birthday", ctl.UpdateBirthday())
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
			response.Error(c, errors.NewBusinessError(constant.ErrInvalidParams))
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
		var req VerifyCodeReq
		if err := c.Bind(&req); err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := ctl.codeSvc.Verify(ctx, req.Biz, req.Phone, req.Code)
		if err != nil {
			response.Error(c, err)
			return
		}

		user, err := ctl.userSvc.FindOrCreateUser(c.Request.Context(), req.Phone)
		if err != nil {
			response.Error(c, err)
			return
		}

		_, tokenErr := ctl.SetLoginToken(c, user.Role, user.Id)
		if tokenErr != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}

		//maskedPhone := req.Phone[:3] + "****" + req.Phone[len(req.Phone)-4:]
		//zap.L().Info(fmt.Sprintf("%s:用户处理成功", req.Biz), zap.String("phone", maskedPhone))

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

		// 检查 Identifier 是否是邮箱
		validate := validator.New()
		isEmail := validate.Var(req.Identifier, "email") == nil

		user, err := ctl.userSvc.Login(c.Request.Context(), req.Identifier, req.Password, isEmail)
		if err != nil {
			response.Error(c, err)
			return
		}

		_, tokenErr := ctl.SetLoginToken(c, user.Role, user.Id)
		if tokenErr != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

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

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}

		claim := claims.(*ijwt.Claims)

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInvalidParams))
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

func (ctl *UserHandler) UpdateBirthday() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateBirthReq
		if err := c.Bind(&req); err != nil {
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		parsedDate, err := time.Parse("2006-01-02", req.Birthday)
		err = ctl.userSvc.UpdateBirthday(c.Request.Context(), claim.Id, parsedDate)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateEmailReq
		if err := c.Bind(&req); err != nil {
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateEmail(c.Request.Context(), claim.Id, req.Email)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateNameReq
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户名:绑定信息错误", zap.Error(err))
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateName(c.Request.Context(), claim.Id, req.Name)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateRoleReq
		if err := c.Bind(&req); err != nil {
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateRole(c.Request.Context(), claim.Id, req.Role)
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
		refreshToken := ctl.ExtractToken(c)
		var rc ijwt.RefreshClaims
		token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = ctl.Handler.CheckSession(c, rc.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置新的 access_token
		_, err = ctl.AccessToken(c, rc.Role, rc.Id, rc.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) LogOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ctl.Handler.ClearToken(c)
		if err != nil {
			response.Error(c, errors.NewBusinessError(constant.ErrInternalServer))
			return
		}

		response.Success(c, nil)
	}
}
