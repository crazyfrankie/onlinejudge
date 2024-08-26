package repository

import (
	"context"
	"database/sql"
	"log"

	"oj/internal/user/domain"
	"oj/internal/user/repository/cache"
	"oj/internal/user/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserDuplicateName  = dao.ErrUserDuplicateName
	ErrUserDuplicatePhone = dao.ErrUserDuplicatePhone
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	UpdateInfo(ctx context.Context, id uint64, user domain.User) error
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

func (ur *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	newUser := dao.User{
		Name:     u.Name,
		Password: u.Password,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: u.Phone,
		Role:  u.Role,
	}
	if err := ur.dao.Insert(ctx, &newUser); err != nil {
		return err
	}
	u.Id = newUser.Id
	return nil
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

func (ur *CacheUserRepository) UpdateInfo(ctx context.Context, id uint64, user domain.User) error {
	// 先更新数据库
	u, err := ur.dao.UpdateInfo(ctx, id, user)
	if err != nil {
		return err
	}

	// 更新缓存中的数据 直接覆盖即可
	go func() {
		newCtx := context.Background()
		err = ur.cache.Set(newCtx, u)
		if err != nil {
			// 记录日志，做监控，但不影响返回的结果
			log.Printf("failed to update cache for user %d: %v", user.Id, err)
		}
	}()

	return err
}
