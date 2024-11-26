package service

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"oj/internal/user/domain"
	"oj/internal/user/repository"
)

var (
	ErrUserNotFound = repository.ErrUserNotFound
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type UserService interface {
	CheckPhone(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateUser(ctx context.Context, phone string) (domain.User, *CodeServiceError)
	Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, *CodeServiceError)
	GetInfo(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WeChatInfo) (domain.User, error)
	FindOrCreateByGithub(ctx context.Context, id int) (domain.User, error)
	UpdateName(ctx context.Context, user domain.User) error
	UpdatePassword(ctx context.Context, user domain.User) error
	UpdateBirthday(ctx context.Context, user domain.User) error
	UpdateEmail(ctx context.Context, user domain.User) error
	UpdateRole(ctx context.Context, user domain.User) error
	GenerateCode() (string, error)
}

type UserSvc struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &UserSvc{
		repo: repo,
	}
}

func (svc *UserSvc) CheckPhone(ctx context.Context, phone string) (domain.User, error) {
	return svc.repo.CheckPhone(ctx, phone)
}

func (svc *UserSvc) FindOrCreateUser(ctx context.Context, phone string) (domain.User, *CodeServiceError) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUserNotFound) {
		return domain.User{}, &CodeServiceError{Code: http.StatusInternalServerError, Message: "system error"}
	}

	var code string
	code, err = svc.GenerateCode()
	if err != nil {
		return domain.User{}, &CodeServiceError{Code: http.StatusInternalServerError, Message: "system error"}
	}
	u := domain.User{
		Phone: phone,
		Name:  code[:15] + "-" + code[15:],
	}

	err = svc.repo.Create(ctx, u)
	if err != nil {
		return domain.User{}, &CodeServiceError{Code: http.StatusInternalServerError, Message: "system error"}
	}

	user, err = svc.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, &CodeServiceError{Code: http.StatusInternalServerError, Message: "system error"}
	}

	return user, nil
}

func (svc *UserSvc) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserSvc) Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, *CodeServiceError) {
	var err error
	var user domain.User

	if isEmail {
		user, err = svc.repo.FindByEmail(ctx, identifier)
	} else {
		user, err = svc.repo.FindByName(ctx, identifier)
	}
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domain.User{}, &CodeServiceError{Code: http.StatusNotFound, Message: "user not found"}
		} else {
			return domain.User{}, &CodeServiceError{Code: http.StatusInternalServerError, Message: "system error"}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 打 DEBUG 日志
		return domain.User{}, &CodeServiceError{Code: http.StatusUnauthorized, Message: "identifier or password error"}
	}

	return user, nil
}

func (svc *UserSvc) GetInfo(ctx context.Context, id uint64) (domain.User, error) {
	return svc.repo.FindByID(ctx, id)
}

func (svc *UserSvc) FindOrCreateByWechat(ctx context.Context, info domain.WeChatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if !errors.Is(err, ErrUserNotFound) {
		return user, nil
	}

	var code string
	code, err = svc.GenerateCode()
	user = domain.User{
		Name:       code[:15] + "-" + code[15:],
		WeChatInfo: info,
	}

	err = svc.repo.Create(ctx, user)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicateWechat) {
		return user, err
	}

	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *UserSvc) FindOrCreateByGithub(ctx context.Context, id int) (domain.User, error) {
	gitId := strconv.Itoa(id)
	user, err := svc.repo.FindByGithub(ctx, gitId)
	if !errors.Is(err, ErrUserNotFound) {
		return user, nil
	}

	var code string
	code, err = svc.GenerateCode()
	user = domain.User{
		Name:     code[:15] + "-" + code[15:],
		GithubID: gitId,
	}

	err = svc.repo.Create(ctx, user)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicateWechat) {
		return user, err
	}

	return svc.repo.FindByGithub(ctx, gitId)
}

func (svc *UserSvc) UpdatePassword(ctx context.Context, user domain.User) error {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashPassword)

	return svc.repo.UpdatePassword(ctx, user)
}

func (svc *UserSvc) UpdateName(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateName(ctx, user)
}

func (svc *UserSvc) UpdateBirthday(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateBirthday(ctx, user)
}

func (svc *UserSvc) UpdateEmail(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateEmail(ctx, user)
}

func (svc *UserSvc) UpdateRole(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateRole(ctx, user)
}

func (svc *UserSvc) GenerateCode() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(20)

	for _, b := range bytes {
		sb.WriteByte(charset[int(b)%len(charset)])
	}

	return sb.String(), nil
}
