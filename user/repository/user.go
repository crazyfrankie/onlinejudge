package repository

import (
	"context"
	"errors"
	"log"
	"oj/user/domain"
	"oj/user/repository/cache"
	"oj/user/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserDuplicateName  = dao.ErrUserDuplicateName
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDao, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error {
	newUser := dao.User{
		Name:     u.Name,
		Password: u.Password,
		Email:    u.Email,
		Role:     u.Role,
	}
	if err := ur.dao.Insert(ctx, newUser); err != nil {
		return err
	}
	u.Id = newUser.Id
	return nil
}

func (ur *UserRepository) FindByName(ctx context.Context, name string) (domain.User, error) {
	user, err := ur.dao.FindByName(ctx, name)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return user, nil
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return user, nil
}

func (ur *UserRepository) FindByID(ctx context.Context, id uint64) (domain.User, error) {
	// 先去缓存中找
	user, err := ur.cache.Get(ctx, id)
	if err == nil {
		// 必然有数据
		return user, nil
	}

	// 去数据库里面加载
	user, err = ur.dao.FindById(ctx, id)
	if errors.Is(err, ErrUserNotFound) {
		return user, ErrUserNotFound
	}
	if err != nil {
		return user, err
	}

	// 异步处理
	go func() {
		// 查询成功后更新缓存
		err = ur.cache.Set(ctx, user)
		if err != nil {
			// 记录日志，做监控，但不影响返回的结果
			log.Printf("failed to update cache for user %d: %v", user.Id, err)
		}
	}()
	return user, err
}
