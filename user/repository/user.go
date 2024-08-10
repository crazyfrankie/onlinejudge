package repository

import (
	"context"

	"oj/user/domain"
	"oj/user/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserDuplicateName  = dao.ErrUserDuplicateName
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
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

func (ur *UserRepository) FindByID(ctx context.Context, id string) (dao.User, error) {
	user, err := ur.dao.FindById(ctx, id)
	if err != nil {
		return dao.User{}, ErrUserNotFound
	}
	return user, nil
}
