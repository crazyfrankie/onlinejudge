package service

import (
	"context"
	"errors"
	"oj/user/repository/dao"
	"oj/user/web/middleware"

	"golang.org/x/crypto/bcrypt"

	"oj/user/domain"
	"oj/user/repository"
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

	var token string
	token, err = middleware.GenerateToken(user.Role, user.Id)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (svc *UserService) GetInfo(ctx context.Context, id string) (dao.User, error) {
	return svc.repo.FindByID(ctx, id)
}
