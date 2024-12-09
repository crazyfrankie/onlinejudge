package web

import (
	"context"
	"fmt"
	"net/http"
	"oj/common/constant"
	"oj/common/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"oj/internal/user/domain"
	ijwt "oj/internal/user/middleware/jwt"
	"oj/internal/user/service"
)

type UserHandler struct {
	userSvc service.UserService
	codeSvc service.CodeService
	logger  *zap.Logger
	ijwt.Handler
}

func NewUserHandler(userSvc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		codeSvc: codeSvc,
		Handler: jwtHdl,
		logger:  logger,
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
		type Req struct {
			Phone string `json:"phone"`
			Biz   string `json:"biz"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error(fmt.Sprintf("%s:绑定参数错误", req.Biz))
			return
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error(fmt.Sprintf("%s:校验参数错误", req.Biz))
			response.Error(c, service.NewBusinessError(constant.ErrInvalidParams))
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
		type Req struct {
			Phone string `json:"phone" validate:"required,len=11"`
			Code  string `json:"code"`
			Role  string `json:"role"`
			Biz   string `json:"biz"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("校验验证码绑定信息错误", zap.Error(err))
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
			zap.L().Error(fmt.Sprintf("%s:查找或创建用户:%s", req.Biz, err.Error()), zap.String("error", err.Error()))
			response.Error(c, err)
			return
		}

		_, tokenErr := ctl.SetLoginToken(c, user.Role, user.Id)
		if tokenErr != nil {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}

		maskedPhone := req.Phone[:3] + "****" + req.Phone[len(req.Phone)-4:]
		zap.L().Info(fmt.Sprintf("%s:用户处理成功", req.Biz), zap.String("phone", maskedPhone))

		response.Success(c, map[string]interface{}{
			"id":    user.Id,
			"phone": user.Phone,
			"name":  user.Name,
		})
	}
}

func (ctl *UserHandler) IdentifierLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Identifier string `json:"identifier"`
			Password   string `json:"password"`
		}
		req := Req{}

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
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}

		response.Success(c, nil)
	}
}

func (ctl *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
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
		type Req struct {
			Password        string `json:"password" validate:"required"`
			ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户密码:绑定信息错误", zap.Error(err))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}

		claim := claims.(*ijwt.Claims)

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("绑定用户信息:信息格式错误", zap.Error(err))
			response.Error(c, service.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		err := ctl.userSvc.UpdatePassword(c.Request.Context(), domain.User{
			Id:       claim.Id,
			Password: req.Password,
		})
		if err != nil {
			zap.L().Error("绑定用户密码:系统错误", zap.Error(err))
			response.Error(c, err)
			return
		}

		zap.L().Info("绑定用户密码成功")
		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateBirthday() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Birthday string `json:"birthday"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户生日:绑定信息错误", zap.Error(err))
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("绑定用户生日:信息格式错误", zap.Error(err))
			response.Error(c, service.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		parsedDate, err := time.Parse("2006-01-02", req.Birthday)
		err = ctl.userSvc.UpdateBirthday(c.Request.Context(), domain.User{
			Id:       claim.Id,
			Birthday: parsedDate,
		})
		if err != nil {
			zap.L().Error("绑定用户生日:系统错误", zap.Error(err))
			response.Error(c, err)
			return
		}

		zap.L().Info("绑定用户生日成功")
		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Email string `json:"email" validate:"required,email"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户邮箱:绑定信息错误", zap.Error(err))
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("绑定用户邮箱:信息格式错误", zap.Error(err))
			response.Error(c, service.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateEmail(c.Request.Context(), domain.User{
			Id:    claim.Id,
			Email: req.Email,
		})
		if err != nil {
			zap.L().Error("绑定用户邮箱:系统错误", zap.Error(err))
			response.Error(c, err)
			return
		}

		zap.L().Info("绑定用户邮箱成功")
		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateName() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name string `json:"name" validate:"required,min=3,max=20"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户名:绑定信息错误", zap.Error(err))
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("绑定用户名:信息格式错误", zap.Error(err))
			response.Error(c, service.NewBusinessError(constant.ErrInvalidParams))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateName(c.Request.Context(), domain.User{
			Id:   claim.Id,
			Name: req.Name,
		})
		if err != nil {
			zap.L().Error("绑定用户名:系统错误", zap.Error(err))
			response.Error(c, err)
			return
		}

		zap.L().Info("绑定用户名成功")
		response.Success(c, nil)
	}
}

func (ctl *UserHandler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Role uint8 `json:"role"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户身份:绑定信息错误", zap.Error(err))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.userSvc.UpdateRole(c.Request.Context(), domain.User{
			Id:   claim.Id,
			Role: req.Role,
		})
		if err != nil {
			zap.L().Error("绑定用户身份:系统错误", zap.Error(err))
			response.Error(c, err)
			return
		}

		zap.L().Info("绑定用户身份成功")
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
			response.Error(c, service.NewBusinessError(constant.ErrInternalServer))
			return
		}

		response.Success(c, nil)
	}
}
