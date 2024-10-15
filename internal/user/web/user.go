package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	svc     service.UserService
	codeSvc service.CodeService
	ijwt.Handler
	cmd redis.Cmdable
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		svc:     svc,
		codeSvc: codeSvc,
		Handler: jwtHdl,
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
			c.JSON(http.StatusTooManyRequests, GetResponse(WithStatus(http.StatusTooManyRequests), WithMsg("send too many")))
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
		default:
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("send successfully")))
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("too many verifications")))
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}
		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("verification successfully")))
	}
}

func (ctl *UserHandler) Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name            string `json:"name" validate:"required,min=3,max=20"`
			Password        string `json:"password" validate:"required,min=8,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789,containsany=$@$!%*#?&"`
			ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
			Email           string `json:"email" validate:"required,email"`
			Phone           string `json:"phone" validate:"required,,regexp=^1[3-9][0-9]{9}$"`
			Role            uint8  `json:"role"`
		}
		req := Req{}
		if err := c.Bind(&req); err != nil {
			return
		}

		// 和之前的自己使用正则表达式校验，在用户名不冲突的情况下，性能几乎一样，只是在可读性上更优
		// 追求代码简洁和可读性时选择此方法，若不追求，根据自己喜好选择
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.svc.Signup(c.Request.Context(), domain.User{
			Name:     req.Name,
			Password: req.Password,
			Email:    req.Email,
			Phone:    req.Phone,
			Role:     req.Role,
		})

		switch {
		case errors.Is(err, service.ErrUserDuplicateEmail):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("email conflict")))
		case errors.Is(err, service.ErrUserDuplicateName):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("name conflict")))
		case errors.Is(err, service.ErrUserDuplicatePhone):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("phone conflict")))
		case err != nil:
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
		default:
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("sign up successfully!")))
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

		// 检查 Identifier 是否是邮箱
		validate := validator.New()
		isEmail := validate.Var(req.Identifier, "email") == nil

		user, err := ctl.svc.Login(c.Request.Context(), req.Identifier, req.Password, isEmail)
		switch {
		case errors.Is(err, service.ErrInvalidUserOrPassword):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("identifier or password error")))
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("identifier not found")))
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("login successfully")))
	}
}

func (ctl *UserHandler) LoginSendSMSCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone" validate:"required,,regexp=^1[3-9][0-9]{9}$"`
		}
		req := Req{}

		if err := c.Bind(&req); err != nil {
			return
		}

		// 手机号格式检测
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("Validation failed: "+err.Error())))
			return
		}

		err := ctl.codeSvc.Send(c.Request.Context(), loginBiz, req.Phone)

		switch {
		case errors.Is(err, service.ErrSendTooMany):
			c.JSON(http.StatusTooManyRequests, GetResponse(WithStatus(http.StatusTooManyRequests), WithMsg("send too many")))
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
		default:
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("send successfully")))
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("too many verifications")))
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		// 设置 JWT
		var user domain.User
		user, err = ctl.svc.FindOrCreate(c.Request.Context(), req.Phone)
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("identifier not found")))
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
			return
		}

		err = ctl.SetLoginToken(c, user.Role, user.Id)
		if err != nil {
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("login successfully")))
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
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}

		claim := claims.(*ijwt.Claims)

		user, err := ctl.svc.GetInfo(c.Request.Context(), claim.Id)

		switch {
		case err != nil:
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
		default:
			c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(User{
				Name:     user.Name,
				Email:    user.Email,
				Phone:    user.Phone,
				Birthday: user.Birthday.Format(time.DateOnly),
				AboutMe:  user.AboutMe,
				Role:     user.Role,
			})))
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("system error")))
			return
		}

		claims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("system error")))
			return
		}
		claim := claims.(*ijwt.Claims)

		err := ctl.svc.EditInfo(c.Request.Context(), claim.Id, domain.User{
			Name:     req.Name,
			Birthday: req.Birthday,
			AboutMe:  req.AboutMe,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetResponse(WithStatus(http.StatusInternalServerError), WithMsg("update failed")))
			return
		}
		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithData(req)))
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
			c.JSON(http.StatusBadRequest, GetResponse(WithStatus(http.StatusBadRequest), WithMsg("log out failed")))
			return
		}

		c.JSON(http.StatusOK, GetResponse(WithStatus(http.StatusOK), WithMsg("log out successfully")))
	}
}
