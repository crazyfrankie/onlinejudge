package web

import (
	"errors"
	"fmt"
	"net/http"
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
		userGroup.POST("send-code", ctl.SendVerificationCode())
		userGroup.POST("verify-code", ctl.VerificationCode())
		userGroup.GET("info", ctl.GetUserInfo())
		userGroup.POST("refresh-token", ctl.TokenRefresh())
	}
}

func (ctl *UserHandler) SendVerificationCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone" validate:"required,len=11"`
			Biz   string `json:"biz"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("发送验证码:绑定信息错误", zap.Error(err))
			return
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("发送验证码:手机号格式校验错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed:"+err.Error())))
			return
		}

		err := ctl.codeSvc.Send(c.Request.Context(), req.Biz, req.Phone)
		switch {
		case errors.Is(err, service.ErrSendTooMany):
			zap.L().Error("发送验证码:发送过于频繁", zap.Error(err))
			c.JSON(http.StatusTooManyRequests, GetResponse(WithStatus(http.StatusTooManyRequests), WithMsg("send too many")))
			return
		case err != nil:
			zap.L().Error("发送验证码:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		default:
			zap.L().Info("发送验证码成功")
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("send successfully")))
		}
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

		_, err := ctl.codeSvc.Verify(c.Request.Context(), req.Biz, req.Phone, req.Code)
		switch {
		case errors.Is(err, service.ErrVerifyTooMany):
			zap.L().Error(fmt.Sprintf("%s:校验验证码:校验次数过多", req.Biz), zap.Error(err))
			c.JSON(http.StatusTooManyRequests, GetResponse(WithStatus(http.StatusTooManyRequests), WithMsg("verify too many")))
			return
		case err != nil:
			zap.L().Error(fmt.Sprintf("%s:校验验证码:系统错误", req.Biz), zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		var user domain.User
		user, err = ctl.userSvc.FindOrCreateUser(c.Request.Context(), req.Phone)
		if err != nil {
			zap.L().Error(fmt.Sprintf("%s:查找或创建用户:系统错误", req.Biz), zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		maskedPhone := req.Phone[:3] + "****" + req.Phone[len(req.Phone)-4:]
		zap.L().Info(fmt.Sprintf("%s:用户处理成功", req.Biz), zap.String("phone", maskedPhone))

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg(fmt.Sprintf("%s successfully", req.Biz)), WithData(map[string]interface{}{
			"id":    user.Id,
			"phone": user.Phone,
			"name":  user.Name,
		})))
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
		switch {
		case errors.Is(err, service.ErrInvalidUserOrPassword):
			c.JSON(http.StatusTooManyRequests, GetResponse(WithStatus(http.StatusTooManyRequests), WithMsg("identifier or password error")))
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, GetResponse(WithStatus(http.StatusNotFound), WithMsg("identifier not found")))
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("login successfully")))
	}
}

func (ctl *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		claim := claims.(*ijwt.Claims)

		user, err := ctl.userSvc.GetInfo(c.Request.Context(), claim.Id)

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
		default:
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(user)))
		}
	}
}

func (ctl *UserHandler) UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId          uint64 `json:"user_id"`
			Password        string `json:"password" validate:"required,min=8,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789,containsany=$@$!%*#?&"`
			ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户密码:绑定信息错误", zap.Error(err))
			return
		}

		// 使用 validator 进行字段验证
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			zap.L().Error("绑定用户信息:信息格式错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.userSvc.UpdatePassword(c.Request.Context(), domain.User{
			Id:       req.UserId,
			Password: req.Password,
		})
		if err != nil {
			zap.L().Error("绑定用户密码:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		zap.L().Info("绑定用户密码成功")

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("bind user's password successfully")))
	}
}

func (ctl *UserHandler) UpdateBirthday() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId   uint64    `json:"user_id"`
			Birthday time.Time `json:"birthday"`
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.userSvc.UpdateBirthday(c.Request.Context(), domain.User{
			Id:       req.UserId,
			Birthday: req.Birthday,
		})
		if err != nil {
			zap.L().Error("绑定用户生日:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		zap.L().Info("绑定用户生日成功")

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("bind user's birthday successfully")))
	}
}

func (ctl *UserHandler) UpdateEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId uint64 `json:"user_id"`
			Email  string `json:"email" validate:"required,email"`
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.userSvc.UpdateBirthday(c.Request.Context(), domain.User{
			Id:    req.UserId,
			Email: req.Email,
		})
		if err != nil {
			zap.L().Error("绑定用户邮箱:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		zap.L().Info("绑定用户邮箱成功")

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("bind user's email successfully")))
	}
}

func (ctl *UserHandler) UpdateName() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId uint64 `json:"user_id"`
			Name   string `json:"name" validate:"required,min=3,max=20"`
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.userSvc.UpdateName(c.Request.Context(), domain.User{
			Id:   req.UserId,
			Name: req.Name,
		})
		if err != nil {
			zap.L().Error("绑定用户名:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		zap.L().Info("绑定用户名成功")

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("bind user's name successfully")))
	}
}

func (ctl *UserHandler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			UserId uint64 `json:"user_id"`
			Role   uint8  `json:"role"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			zap.L().Error("绑定用户身份:绑定信息错误", zap.Error(err))
			return
		}

		err := ctl.userSvc.UpdateName(c.Request.Context(), domain.User{
			Id:   req.UserId,
			Role: req.Role,
		})
		if err != nil {
			zap.L().Error("绑定用户身份:系统错误", zap.Error(err))
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		zap.L().Info("绑定用户身份成功")

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("bind user's role successfully")))
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
		err = ctl.AccessToken(c, rc.Role, rc.Id, rc.SSId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("refresh successfully")))
	}
}

func (ctl *UserHandler) LogOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ctl.Handler.ClearToken(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("log out failed")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("log out successfully")))
	}
}
