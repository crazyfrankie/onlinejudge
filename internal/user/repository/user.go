package repository

import (
	"context"
	"log"
	"time"

	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/user/repository/dao"
)

var (
	ErrUserNotFound        = dao.ErrUserNotFound
	ErrUserDuplicateWechat = dao.ErrUserDuplicateWechat
	ErrUserDuplicateGithub = dao.ErrUserDuplicateGithub
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	CreateByGithub(ctx context.Context, user domain.User) error
	CreateByWeChat(ctx context.Context, user domain.User) error
	CheckPhone(ctx context.Context, phone string) (domain.User, error)
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	FindByGithub(ctx context.Context, gitId string) (domain.User, error)
	UpdatePassword(ctx context.Context, uid uint64, password string) error
	UpdateBirthday(ctx context.Context, uid uint64, birth time.Time) error
	UpdateName(ctx context.Context, uid uint64, name string) error
	UpdateEmail(ctx context.Context, uid uint64, email string) error
	UpdateRole(ctx context.Context, uid uint64, role uint8) error
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

func (ur *CacheUserRepository) CheckPhone(ctx context.Context, phone string) (domain.User, error) {
	return ur.dao.FindByPhone(ctx, phone)
}

func (ur *CacheUserRepository) Create(ctx context.Context, user domain.User) error {
	return ur.dao.Insert(ctx, user)
}

func (ur *CacheUserRepository) CreateByGithub(ctx context.Context, user domain.User) error {
	return ur.dao.InsertByGithub(ctx, user)
}

func (ur *CacheUserRepository) CreateByWeChat(ctx context.Context, user domain.User) error {
	return ur.dao.InsertByWeChat(ctx, user)
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

	// 回写缓存
	// 可异步也可以同步，看场景的
	err = ur.cache.Set(ctx, user)
	if err != nil {
		// 记录日志，做监控，但不影响返回的结果
		log.Printf("failed to update cache for user %d: %v", user.Id, err)
	}

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

func (ur *CacheUserRepository) UpdatePassword(ctx context.Context, uid uint64, password string) error {
	updateUser, err := ur.dao.UpdatePassword(ctx, uid, password)
	if err != nil {
		return err
	}

	// 删除缓存
	cacheErr := ur.cache.Del(ctx, updateUser.Id)
	if cacheErr != nil {
		log.Printf("failed to update cache for user %d: %v", updateUser.Id, cacheErr)
	}

	return nil
}

func (ur *CacheUserRepository) UpdateBirthday(ctx context.Context, uid uint64, birth time.Time) error {
	updateUser, err := ur.dao.UpdateBirthday(ctx, uid, birth)
	if err != nil {
		return err
	}

	// 删除缓存
	cacheErr := ur.cache.Del(ctx, updateUser.Id)
	if cacheErr != nil {
		log.Printf("failed to update cache for user %d: %v", uid, cacheErr)
	}

	return nil
}

func (ur *CacheUserRepository) UpdateName(ctx context.Context, uid uint64, name string) error {
	updateUser, err := ur.dao.UpdateName(ctx, uid, name)
	if err != nil {
		return err
	}

	// 删除缓存
	cacheErr := ur.cache.Del(ctx, updateUser.Id)
	if cacheErr != nil {
		log.Printf("failed to update cache for user %d: %v", uid, cacheErr)
	}

	return nil
}

func (ur *CacheUserRepository) UpdateEmail(ctx context.Context, uid uint64, email string) error {
	// 更新数据库
	updateUser, err := ur.dao.UpdateEmail(ctx, uid, email)
	if err != nil {
		return err
	}

	// 删除缓存
	cacheErr := ur.cache.Del(ctx, updateUser.Id)
	if cacheErr != nil {
		log.Printf("failed to update cache for user %d: %v", uid, cacheErr)
	}

	return nil
}

func (ur *CacheUserRepository) UpdateRole(ctx context.Context, uid uint64, role uint8) error {
	updateUser, err := ur.dao.UpdateRole(ctx, uid, role)
	if err != nil {
		return err
	}

	// 删除缓存
	cacheErr := ur.cache.Del(ctx, updateUser.Id)
	if cacheErr != nil {
		log.Printf("failed to update cache for user %d: %v", uid, cacheErr)
	}

	return nil
}
