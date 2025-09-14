package service

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	
	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository"
)

var (
	ErrUserNotFound = repository.ErrUserNotFound
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type UserService interface {
	FindOrCreateUser(ctx context.Context, phone string) (domain.User, error)
	Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, error)
	GetInfo(ctx context.Context, id uint64) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WeChatInfo) (domain.User, error)
	FindOrCreateByGithub(ctx context.Context, id int) (domain.User, error)
	UpdatePassword(ctx context.Context, uid uint64, password string) error
	UpdateInfo(ctx context.Context, u domain.User) error
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

func (svc *UserSvc) FindOrCreateUser(ctx context.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUserNotFound) {
		return domain.User{}, er.NewBizError(constant.ErrUserNotFound)
	}

	var code string
	code, err = svc.GenerateCode()
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{
		Phone: phone,
		Name:  code[:15] + "-" + code[15:],
	}

	err = svc.repo.Create(ctx, u)
	if err != nil {
		return domain.User{}, er.NewBizError(constant.ErrUserInternalServer)
	}

	user, err = svc.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, er.NewBizError(constant.ErrUserInternalServer)
	}

	return user, nil
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
			return domain.User{}, er.NewBizError(constant.ErrUserNotFound)
		} else {
			return domain.User{}, er.NewBizError(constant.ErrUserInternalServer)
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 打 DEBUG 日志
		return domain.User{}, er.NewBizError(constant.ErrInvalidCredentials)
	}

	return user, nil
}

func (svc *UserSvc) GetInfo(ctx context.Context, id uint64) (domain.User, error) {
	user, err := svc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domain.User{}, er.NewBizError(constant.ErrUserNotFound)
		}

		return domain.User{}, er.NewBizError(constant.ErrUserInternalServer)
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
		return user, er.NewBizError(constant.ErrUserInternalServer)
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
		return user, er.NewBizError(constant.ErrUserInternalServer)
	}

	return svc.repo.FindByGithub(ctx, gitId)
}

func (svc *UserSvc) UpdatePassword(ctx context.Context, uid uint64, password string) error {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return er.NewBizError(constant.ErrUserInternalServer)
	}
	password = string(hashPassword)

	err = svc.repo.UpdatePassword(ctx, uid, password)
	if err != nil {
		return er.NewBizError(constant.ErrUserInternalServer)
	}

	return nil
}

func (svc *UserSvc) UpdateInfo(ctx context.Context, u domain.User) error {
	fds := updateFds(u)

	err := svc.repo.UpdateInfo(ctx, u.Id, fds)
	if err != nil {
		return er.NewBizError(constant.ErrInternalServer)
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

func updateFds(u domain.User) map[string]any {
	res := make(map[string]any)
	if u.Birthday != (time.Time{}) {
		res["birthday"] = u.Birthday
	}
	if u.Email != "" {
		res["email"] = u.Email
	}
	if u.Name != "" {
		res["name"] = u.Name
	}

	return res
}
