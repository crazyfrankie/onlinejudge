package web

import (
	"context"
	"errors"
	"net/http"
	"oj/user/service/biz"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"oj/user/domain"
	"oj/user/service"
)

type UserHandler struct {
	svc              *service.UserService
	codeSvc          *biz.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService, codeSvc *biz.CodeService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-zA-Z])(?=.*\d)(?=.*[$@$!%*#?&])[a-zA-Z\d$@$!%*#?&]{8,}$`
	)
	emailRegexExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordRegexExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    emailRegexExp,
		passwordRegexExp: passwordRegexExp,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("user")
	{
		userGroup.POST("/signup", ctl.Signup())
		userGroup.POST("/login", ctl.Login())
		userGroup.POST("/logout", ctl.Logout())
		userGroup.GET("/:id", ctl.GetInfo())
		userGroup.POST("/login_sms/code/send", ctl.LoginSendSMSCode())
		userGroup.POST("/login_sms", ctl.LoginVerifySMSCode())
	}
}

func (ctl *UserHandler) Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name            string `json:"name"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
			Email           string `json:"email"`
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
			c.JSON(http.StatusBadRequest, "system error 1")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your email does not fit the format")
			return
		}

		// 密码格式检查
		ok, err = ctl.passwordRegexExp.MatchString(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error 2")
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, "your password does not fit the format")
			return
		}

		err = ctl.svc.Signup(c.Request.Context(), domain.User{
			Name:     req.Name,
			Password: req.Password,
			Email:    req.Email,
			Role:     req.Role,
		})
		if errors.Is(err, service.ErrUserDuplicateEmail) {
			c.JSON(http.StatusInternalServerError, "email conflict")
			return
		}
		if errors.Is(err, service.ErrUserDuplicateName) {
			c.JSON(http.StatusInternalServerError, "name conflict")
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		c.JSON(http.StatusOK, "sign up successfully!")
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
		if errors.Is(err, service.ErrInvalidUserOrPassword) {
			c.JSON(http.StatusInternalServerError, "identifier or password error")
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusInternalServerError, "identifier not found")
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		c.Header("x-jwt-token", token)

		c.JSON(http.StatusOK, "login successfully!")
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

		const Biz = "login"
		err := ctl.codeSvc.Send(c.Request.Context(), Biz, req.Phone)
		if errors.Is(err, biz.ErrSendTooMany) {
			c.JSON(http.StatusTooManyRequests, "send too many")
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}
		c.JSON(http.StatusOK, "send successfully")
	}
}

func (ctl *UserHandler) LoginVerifySMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctl.codeSvc.Verify(c.Request.Context())
	}
}

func (ctl *UserHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		sess.Options(sessions.Options{
			// 退出登录
			MaxAge: -1,
		})
		c.JSON(http.StatusOK, "log out successfully!")
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
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
