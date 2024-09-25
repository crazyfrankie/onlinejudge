package service

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"oj/internal/middleware"
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

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, error)
	GetInfo(ctx context.Context, id uint64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WeChatInfo) (domain.User, error)
	EditInfo(ctx context.Context, id uint64, user domain.User) error
	GenerateCode() string
}

type UserSvc struct {
	repo   repository.UserRepository
	jwtGen middleware.TokenGenerator
}

func NewUserService(repo repository.UserRepository, jwtGen middleware.TokenGenerator) UserService {
	return &UserSvc{
		repo:   repo,
		jwtGen: jwtGen,
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
	// 明确知道，没有这个用户
	code := svc.GenerateCode()
	user = domain.User{
		Name:  "用户" + code,
		Phone: phone,
	}
	err = svc.repo.Create(ctx, user)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicatePhone) {
		return user, err
	}
	// 有主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserSvc) FindOrCreateByWechat(ctx context.Context, info domain.WeChatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if errors.Is(err, ErrUserNotFound) {
		return user, nil
	}

	code := svc.GenerateCode()
	user = domain.User{
		Name:       "用户" + code,
		WeChatInfo: info,
	}

	err = svc.repo.Create(ctx, user)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicateWechat) {
		return user, err
	}

	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *UserSvc) EditInfo(ctx context.Context, id uint64, user domain.User) error {
	// 可以考虑删除，因为整体业务逻辑保证了这个用户一定存在
	existingUser, err := svc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 只有非空字段才会更新
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.AboutMe != "" {
		existingUser.AboutMe = user.AboutMe
	}
	if !user.Birthday.IsZero() {
		existingUser.Birthday = user.Birthday
	}

	return svc.repo.UpdateInfo(ctx, id, existingUser)
}

func (svc *UserSvc) GenerateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var code strings.Builder
	for i := 0; i < 6; i++ {
		digit := rand.Intn(10)
		code.WriteString(strconv.Itoa(digit))
	}
	return code.String()
}
