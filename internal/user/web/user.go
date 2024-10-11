package web

import (
	"errors"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"

	"oj/internal/user/domain"
	"oj/internal/user/service"
	ijwt "oj/internal/user/web/jwt"
)

const (
	signUpBiz = "signup"
	loginBiz  = "login"
)

type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	phoneRegexExp    *regexp.Regexp
	ijwt.Handler
	cmd redis.Cmdable
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler) *UserHandler {
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
		Handler:          jwtHdl,
		emailRegexExp:    emailRegexExp,
		passwordRegexExp: passwordRegexExp,
		phoneRegexExp:    phoneRegexExp,
	}
}

func (ctl *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("user")
	{
		userGroup.POST("signup/send-code", ctl.SignupSendSMSCode())
		userGroup.POST("signup/verify-code", ctl.SignupVerifySMSCode())
		userGroup.POST("signup", ctl.Signup())
		userGroup.POST("login", ctl.Login())
		userGroup.POST("login/send-code", ctl.LoginSendSMSCode())
		userGroup.POST("login-sms", ctl.LoginVerifySMSCode())
		userGroup.GET("info", ctl.GetUserInfo())
		userGroup.POST("info/edit", ctl.EditUserInfo())
		userGroup.POST("refresh_token", ctl.TokenRefresh())
	}
}

func (ctl *UserHandler) SignupSendSMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		phone := c.PostForm("phone")

		err := ctl.codeSvc.Send(c.Request.Context(), signUpBiz, phone)
		switch {
		case errors.Is(err, service.ErrSendTooMany):
			c.JSON(http.StatusTooManyRequests, "send too many")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, "send successfully")
		}
	}
}

func (ctl *UserHandler) SignupVerifySMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		phone := c.PostForm("phone")
		code := c.PostForm("code")

		_, err := ctl.codeSvc.Verify(c.Request.Context(), signUpBiz, phone, code)
		switch {
		case errors.Is(err, service.ErrVerifyTooMany):
			c.JSON(http.StatusBadRequest, "too many verifications")
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}
		c.JSON(http.StatusOK, "verification successfully")
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
			c.JSON(http.StatusBadRequest, "password does not match")
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
		case errors.Is(err, service.ErrUserDuplicateEmail):
			c.JSON(http.StatusInternalServerError, "email conflict")
		case errors.Is(err, service.ErrUserDuplicateName):
			c.JSON(http.StatusInternalServerError, "name conflict")
		case errors.Is(err, service.ErrUserDuplicatePhone):
			c.JSON(http.StatusInternalServerError, "phone conflict")
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
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

		var user domain.User
		user, err = ctl.svc.Login(c.Request.Context(), req.Identifier, req.Password, isEmail)
		switch {
		case errors.Is(err, service.ErrInvalidUserOrPassword):
			c.JSON(http.StatusInternalServerError, "identifier or password error")
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, "identifier not found")
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		c.JSON(http.StatusOK, "login successfully")
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

		err = ctl.codeSvc.Send(c.Request.Context(), loginBiz, req.Phone)

		switch {
		case errors.Is(err, service.ErrSendTooMany):
			c.JSON(http.StatusTooManyRequests, "send too many")
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
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

		_, err := ctl.codeSvc.Verify(c.Request.Context(), loginBiz, req.Phone, req.Code)
		switch {
		case errors.Is(err, service.ErrVerifyTooMany):
			c.JSON(http.StatusBadRequest, "too many verifications")
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		// 设置 JWT
		var user domain.User
		user, err = ctl.svc.FindOrCreate(c.Request.Context(), req.Phone)
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, "identifier not found")
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		c.JSON(http.StatusOK, "login successfully")
	}
}

func (ctl *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		type User struct {
			Name     string
			Email    string
			Phone    string
			Birthday string
			AboutMe  string
			Role     uint8
		}

		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}

		claim := claims.(*ijwt.Claims)

		user, err := ctl.svc.GetInfo(c.Request.Context(), claim.Id)

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, "system error")
		default:
			c.JSON(http.StatusOK, User{
				Name:     user.Name,
				Email:    user.Email,
				Phone:    user.Phone,
				Birthday: user.Birthday.Format(time.DateOnly),
				AboutMe:  user.AboutMe,
				Role:     user.Role,
			})
		}
	}
}

func (ctl *UserHandler) EditUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name     string    `json:"name"`
			Birthday time.Time `json:"birthday"`
			AboutMe  string    `json:"aboutMe"`
		}

		var req Req
		if err := c.Bind(&req); err != nil {
			c.JSON(http.StatusBadRequest, "system error")
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, "system error")
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.svc.EditInfo(c.Request.Context(), claim.Id, domain.User{
			Name:     req.Name,
			Birthday: req.Birthday,
			AboutMe:  req.AboutMe,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "update failed")
			return
		}
		c.JSON(http.StatusOK, req)
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

		c.JSON(http.StatusOK, "refresh successfully")
	}
}

func (ctl *UserHandler) LogOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ctl.Handler.ClearToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, "log out failed")
			return
		}

		c.JSON(http.StatusOK, "log out successfully")
	}
}
