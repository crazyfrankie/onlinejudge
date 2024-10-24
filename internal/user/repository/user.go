package repository

import (
	"context"
	"log"

	"oj/internal/user/domain"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
)

var (
	ErrUserDuplicateEmail  = dao.ErrUserDuplicateEmail
	ErrUserDuplicateName   = dao.ErrUserDuplicateName
	ErrUserDuplicatePhone  = dao.ErrUserDuplicatePhone
	ErrUserNotFound        = dao.ErrUserNotFound
	ErrUserDuplicateWechat = dao.ErrUserDuplicateWechat
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	CheckPhone(ctx context.Context, phone string) error
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	FindByGithub(ctx context.Context, gitId string) (domain.User, error)
	UpdatePassword(ctx context.Context, user domain.User) error
	UpdateBirthday(ctx context.Context, user domain.User) error
	UpdateName(ctx context.Context, user domain.User) error
	UpdateEmail(ctx context.Context, user domain.User) error
	UpdateRole(ctx context.Context, user domain.User) error
}

type CacheUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (ur *CacheUserRepository) CheckPhone(ctx context.Context, phone string) error {
	_, err := ur.dao.FindByPhone(ctx, phone)
	return err
}

func (ur *CacheUserRepository) Create(ctx context.Context, user domain.User) error {
	return ur.dao.Insert(ctx, user)
}

func (ur *CacheUserRepository) FindByName(ctx context.Context, name string) (domain.User, error) {
	user, err := ur.dao.FindByName(ctx, name)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (ur *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (ur *CacheUserRepository) FindByID(ctx context.Context, id uint64) (domain.User, error) {
	// 先去缓存中找
	user, err := ur.cache.Get(ctx, id)
	if err == nil {
		// 必然有数据
		return user, nil
	}

	// 去数据库里面加载
	user, err = ur.dao.FindById(ctx, id)
	if err != nil {
		return user, err
	}

	// 异步处理
	go func() {
		newCtx := context.Background()
		// 查询成功后更新缓存
		err = ur.cache.Set(newCtx, user)
		if err != nil {
			// 记录日志，做监控，但不影响返回的结果
			log.Printf("failed to update cache for user %d: %v", user.Id, err)
		}
	}()
	return user, err
}

func (ur *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return user, err
}

func (ur *CacheUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	user, err := ur.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return user, err
}

func (ur *CacheUserRepository) FindByGithub(ctx context.Context, gitId string) (domain.User, error) {
	return ur.dao.FindByGithub(ctx, gitId)
}

func (ur *CacheUserRepository) UpdatePassword(ctx context.Context, user domain.User) error {
	return ur.dao.UpdatePassword(ctx, user)
}

func (ur *CacheUserRepository) UpdateBirthday(ctx context.Context, user domain.User) error {
	return ur.dao.UpdateBirthday(ctx, user)
}

func (ur *CacheUserRepository) UpdateName(ctx context.Context, user domain.User) error {
	return ur.dao.UpdateName(ctx, user)
}

func (ur *CacheUserRepository) UpdateEmail(ctx context.Context, user domain.User) error {
	return ur.dao.UpdateEmail(ctx, user)
}

func (ur *CacheUserRepository) UpdateRole(ctx context.Context, user domain.User) error {
	return ur.dao.UpdateRole(ctx, user)
}
