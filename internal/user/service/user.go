package service

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"oj/internal/user/domain"
	"oj/internal/user/repository"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserDuplicateName     = repository.ErrUserDuplicateName
	ErrUserDuplicatePhone    = repository.ErrUserDuplicatePhone
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvalidUserOrPassword = errors.New("identifier or password error")
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type UserService interface {
	CheckPhone(ctx context.Context, phone string) error
	CreateUser(ctx context.Context, user domain.User) error
	Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, error)
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

func (svc *UserSvc) CheckPhone(ctx context.Context, phone string) error {
	return svc.repo.CheckPhone(ctx, phone)
}

func (svc *UserSvc) CreateUser(ctx context.Context, user domain.User) error {
	err := svc.repo.Create(ctx, user)
	if err != nil {
		return err
	}

	var code string
	code, err = svc.GenerateCode()
	user = domain.User{
		Name: code[:15] + "-" + code[15:],
	}

	return nil
}

func (svc *UserSvc) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserSvc) Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, error) {
	var err error
	var user domain.User

	if isEmail {
		user, err = svc.repo.FindByEmail(ctx, identifier)
	} else {
		user, err = svc.repo.FindByName(ctx, identifier)
	}
	if err != nil {
		return domain.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 打 DEBUG 日志
		return domain.User{}, ErrInvalidUserOrPassword
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
