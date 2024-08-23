package service

import (
	"context"
	"encoding/gob"
	"errors"
	"github.com/golang-jwt/jwt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"oj/user/domain"
	"oj/user/repository"
	"oj/user/web/middleware"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserDuplicateName     = repository.ErrUserDuplicateName
	ErrUserDuplicatePhone    = repository.ErrUserDuplicatePhone
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvalidUserOrPassword = errors.New("identifier or password error")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, identifier, password string, isEmail bool) (string, error)
	GenerateToken(role uint8, id uint64, userAgent string) (string, error)
	GetInfo(ctx context.Context, id string) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type UserSvc struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &UserSvc{
		repo: repo,
	}
}

func (svc *UserSvc) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserSvc) Login(ctx context.Context, identifier, password string, isEmail bool) (string, error) {
	var err error
	var user domain.User

	if isEmail {
		user, err = svc.repo.FindByEmail(ctx, identifier)
	} else {
		user, err = svc.repo.FindByName(ctx, identifier)
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 打 DEBUG 日志
		return "", ErrInvalidUserOrPassword
	}

	// 从上下文中取出 UserAgent
	userAgent := ctx.Value("UserAgent").(string)

	var token string
	token, err = svc.GenerateToken(user.Role, user.Id, userAgent)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (svc *UserSvc) GenerateToken(role uint8, id uint64, userAgent string) (string, error) {
	gob.Register(time.Now())
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)
	claims := middleware.Claims{
		Role: role,
		Id:   id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "oj",
		},
		UserAgent: userAgent,
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(middleware.SecretKey)
	return token, err
}

func (svc *UserSvc) GetInfo(ctx context.Context, id string) (domain.User, error) {
	Id, err := strconv.Atoi(id)
	if err != nil {
		return domain.User{}, err
	}
	return svc.repo.FindByID(ctx, uint64(Id))
}

func (svc *UserSvc) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	// 快路径
	if !errors.Is(err, ErrUserNotFound) {
		// 绝大部分请求都会进来这里
		return user, nil
	}
	// 在系统资源不足后，触发降级策略
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}

	// 慢路径
	// 你明确知道，没有这个用户
	user = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, user)
	if err != nil || !errors.Is(err, repository.ErrUserDuplicatePhone) {
		return user, err
	}
	// 有主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}
