package service

import (
	"context"
	"errors"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"oj/user/domain"
	"oj/user/repository"
	"oj/user/web/middleware"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserDuplicateName     = repository.ErrUserDuplicateName
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvalidUserOrPassword = errors.New("identifier or password error")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, identifier, password string, isEmail bool) (string, error) {
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
	token, err = middleware.GenerateToken(user.Role, user.Id, userAgent)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (svc *UserService) GetInfo(ctx context.Context, id string) (domain.User, error) {
	Id, err := strconv.Atoi(id)
	if err != nil {
		return domain.User{}, err
	}
	return svc.repo.FindByID(ctx, uint64(Id))
}
