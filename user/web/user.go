package web

import (
	"context"
	"errors"
	"net/http"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"

	"oj/user/domain"
	"oj/user/service"
	"oj/user/service/biz"
)

const Biz = "login"

type UserHandler struct {
	svc              *service.UserService
	codeSvc          *biz.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	phoneRegexExp    *regexp.Regexp
}

func NewUserHandler(svc *service.UserService, codeSvc *biz.CodeService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-zA-Z])(?=.*\d)(?=.*[$@$!%*#?&])[a-zA-Z\d$@$!%*#?&]{8,}$`
		phoneRegexPattern    = `^1[3-9]\d{9}$`
	)
	emailRegexExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordRegexExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	phoneRegexExp := regexp.MustCompile(phoneRegexPattern, regexp.None)
	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    emailRegexExp,
		passwordRegexExp: passwordRegexExp,
		phoneRegexExp:    phoneRegexExp,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("user")
	{
		userGroup.POST("/signup", ctl.Signup())
		userGroup.POST("/login", ctl.Login())
		userGroup.GET("/:id", ctl.GetInfo())
		userGroup.POST("/login_sms/code/send", ctl.LoginSendSMSCode())
		userGroup.POST("/sms_login", ctl.LoginVerifySMSCode())
	}
}

func (ctl *UserHandler) Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name            string `json:"name"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
			Email           string `json:"email"`
			Phone           string `json:"phone"`
			Role            uint8  `json:"role"`
		}
		req := Req{}
		if err := c.Bind(&req); err != nil {
			return
		}

		// 两次密码不一致
		if req.Password != req.ConfirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
			return
		}

		// 邮箱格式检查
		ok, err := ctl.emailRegexExp.MatchString(req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your email does not fit the format")
			return
		}

		// 密码格式检查
		ok, err = ctl.passwordRegexExp.MatchString(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your password does not fit the format")
			return
		}

		// 手机号格式检测
		ok, err = ctl.phoneRegexExp.MatchString(req.Phone)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your phone number does not fit the format")
			return
		}

		err = ctl.svc.Signup(c.Request.Context(), domain.User{
			Name:     req.Name,
			Password: req.Password,
			Email:    req.Email,
			Phone:    req.Phone,
			Role:     req.Role,
		})

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		case errors.Is(err, service.ErrUserDuplicateEmail):
			c.JSON(http.StatusInternalServerError, "email conflict")
		case errors.Is(err, service.ErrUserDuplicateName):
			c.JSON(http.StatusInternalServerError, "name conflict")
		default:
			c.JSON(http.StatusOK, "sign up successfully!")
		}
	}
}

func (ctl *UserHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Identifier string `json:"identifier"`
			Password   string `json:"password"`
		}
		req := Req{}

		if err := c.Bind(&req); err != nil {
			return
		}

		// 检查是否包含邮箱
		isEmail, err := ctl.emailRegexExp.MatchString(req.Identifier)

		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		// 创建带有UserAgent的context.Context
		ctx := context.WithValue(c.Request.Context(), "UserAgent", c.Request.UserAgent())

		var token string
		token, err = ctl.svc.Login(ctx, req.Identifier, req.Password, isEmail)
		switch {
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
		case errors.Is(err, service.ErrInvalidUserOrPassword):
			c.JSON(http.StatusInternalServerError, "identifier or password error")
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, "identifier not found")
		default:
			c.Header("x-jwt-token", token)
			c.JSON(http.StatusOK, "login successfully!")
		}
	}
}

func (ctl *UserHandler) LoginSendSMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone"`
		}
		req := Req{}

		if err := c.Bind(&req); err != nil {
			return
		}

		// 手机号格式检测
		ok, err := ctl.phoneRegexExp.MatchString(req.Phone)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your phone number does not fit the format")
			return
		}

		err = ctl.codeSvc.Send(c.Request.Context(), Biz, req.Phone)

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		case errors.Is(err, biz.ErrSendTooMany):
			c.JSON(http.StatusTooManyRequests, "send too many")
		default:
			c.JSON(http.StatusOK, "send successfully")
		}
	}
}

func (ctl *UserHandler) LoginVerifySMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone"`
			Code  string `json:"code"`
		}
		req := Req{}

		if err := c.Bind(&req); err != nil {
			return
		}

		ok, err := ctl.codeSvc.Verify(c.Request.Context(), Biz, req.Phone, req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "verify code error")
			return
		}

		// 设置 JWT
		var user domain.User
		user, err = ctl.svc.FindOrCreate(c.Request.Context(), req.Phone)
		switch {
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, "identifier not found")
			return
		}

		// 从 Header 中取出 UserAgent
		userAgent := c.GetHeader("User-Agent")

		var token string
		token, err = ctl.svc.GenerateToken(user.Role, user.Id, userAgent)
		switch {
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
		default:
			c.Header("x-jwt-token", token)
			c.JSON(http.StatusOK, "login successfully")
		}
	}
}

func (ctl *UserHandler) GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, "unable to get id")
			return
		}

		user, err := ctl.svc.GetInfo(c.Request.Context(), id)

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, "user not found")
		default:
			c.JSON(http.StatusOK, user)
		}
	}
}
