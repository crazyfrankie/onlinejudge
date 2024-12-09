package service

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"oj/common/constant"
	"oj/internal/user/domain"
	"oj/internal/user/repository"
)

var (
	ErrUserNotFound = repository.ErrUserNotFound
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type UserService interface {
	CheckPhone(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateUser(ctx context.Context, phone string) (domain.User, error)
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

func (svc *UserSvc) CheckPhone(ctx context.Context, phone string) (domain.User, error) {
	return svc.repo.CheckPhone(ctx, phone)
}

func (svc *UserSvc) FindOrCreateUser(ctx context.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUserNotFound) {
		return domain.User{}, NewBusinessError(constant.ErrUserNotFound)
	}

	var code string
	code, err = svc.GenerateCode()
	if err != nil {
		return domain.User{}, NewBusinessError(constant.ErrInternalServer)
	}
	u := domain.User{
		Phone: phone,
		Name:  code[:15] + "-" + code[15:],
	}

	err = svc.repo.Create(ctx, u)
	if err != nil {
		return domain.User{}, NewBusinessError(constant.ErrInternalServer)
	}

	user, err = svc.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, NewBusinessError(constant.ErrInternalServer)
	}

	return user, nil
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
		if errors.Is(err, ErrUserNotFound) {
			return domain.User{}, NewBusinessError(constant.ErrUserNotFound)
		} else {
			return domain.User{}, NewBusinessError(constant.ErrInternalServer)
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 打 DEBUG 日志
		return domain.User{}, NewBusinessError(constant.ErrInvalidCredentials)
	}

	return user, nil
}

func (svc *UserSvc) GetInfo(ctx context.Context, id uint64) (domain.User, error) {
	user, err := svc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domain.User{}, NewBusinessError(constant.ErrUserNotFound)
		}

		return domain.User{}, NewBusinessError(constant.ErrInternalServer)
	}

	return user, nil
}

func (svc *UserSvc) FindOrCreateByWechat(ctx context.Context, info domain.WeChatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if err == nil {
		return user, nil
	}

	var code string
	code, err = svc.GenerateCode()
	user = domain.User{
		Name:       code[:15] + "-" + code[15:],
		WeChatInfo: info,
	}

	err = svc.repo.CreateByWeChat(ctx, user)
	if err != nil {
		return user, nil
	}

	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *UserSvc) FindOrCreateByGithub(ctx context.Context, id int) (domain.User, error) {
	gitId := strconv.Itoa(id)
	user, err := svc.repo.FindByGithub(ctx, gitId)
	if err == nil {
		return user, nil
	}

	var code string
	code, err = svc.GenerateCode()
	user = domain.User{
		Name:     code[:15] + "-" + code[15:],
		GithubID: gitId,
	}

	err = svc.repo.CreateByGithub(ctx, user)
	if err != nil {
		return user, nil
	}

	return svc.repo.FindByGithub(ctx, gitId)
}

func (svc *UserSvc) UpdatePassword(ctx context.Context, user domain.User) error {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}
	user.Password = string(hashPassword)

	err = svc.repo.UpdatePassword(ctx, user)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *UserSvc) UpdateName(ctx context.Context, user domain.User) error {
	err := svc.repo.UpdateName(ctx, user)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *UserSvc) UpdateBirthday(ctx context.Context, user domain.User) error {
	err := svc.repo.UpdateBirthday(ctx, user)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *UserSvc) UpdateEmail(ctx context.Context, user domain.User) error {
	err := svc.repo.UpdateEmail(ctx, user)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}

	return nil
}

func (svc *UserSvc) UpdateRole(ctx context.Context, user domain.User) error {
	err := svc.repo.UpdateRole(ctx, user)
	if err != nil {
		return NewBusinessError(constant.ErrInternalServer)
	}

	return nil
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
