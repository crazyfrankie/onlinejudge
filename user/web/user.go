package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"

	"oj/user/domain"
	"oj/user/service"
)

type UserHandler struct {
	svc              *service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-zA-Z])(?=.*\d)(?=.*[$@$!%*#?&])[a-zA-Z\d$@$!%*#?&]{8,}$`
	)
	emailRegexExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordRegexExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:              svc,
		emailRegexExp:    emailRegexExp,
		passwordRegexExp: passwordRegexExp,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("user")
	{
		userGroup.POST("/signup", ctl.Signup())
		userGroup.POST("/login", ctl.Login())
	}
}

func (ctl *UserHandler) Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name            string `json:"name"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
			Email           string `json:"email"`
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

		var token string
		token, err = ctl.svc.Login(c.Request.Context(), req.Identifier, req.Password, isEmail)
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
func (ctl *UserHandler) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Options(sessions.Options{
		// 退出登录
		MaxAge: -1,
	})
	c.JSON(http.StatusOK, "log out successfully!")
}
