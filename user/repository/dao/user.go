package dao

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"oj/user/domain"
)

type UserDao struct {
	db *gorm.DB
}

type User struct {
	Id       uint64 `gorm:"primaryKey,autoIncrement"`
	Name     string `gorm:"unique;not null"`
	Password string
	Email    string `gorm:"unique;not null"`
	Role     uint8
	Ctime    int64
	Uptime   int64
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

var (
	ErrUserDuplicateEmail = errors.New("duplicate email")
	ErrUserDuplicateName  = errors.New("duplicate name")
	ErrUserNotFound       = errors.New("user not found")
)

func handleDBError(err error) error {
	var mysqlErr *mysql.MySQLError
	const uniqueConflictErrNo uint16 = 1062

	if errors.As(err, &mysqlErr) && mysqlErr.Number == uniqueConflictErrNo {
		if strings.Contains(mysqlErr.Message, "email") {
			return ErrUserDuplicateEmail
		} else if strings.Contains(mysqlErr.Message, "name") {
			return ErrUserDuplicateName
		}
	}
	return err
}

func (dao *UserDao) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	return handleDBError(err)
}

func (dao *UserDao) FindByName(ctx context.Context, name string) (domain.User, error) {
	var user domain.User

	result := dao.db.WithContext(ctx).Where("name = ?", name).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}
	return user, nil
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User

	result := dao.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}
	return user, nil
}

func (dao *UserDao) FindById(ctx context.Context, id uint64) (domain.User, error) {
	var user User

	result := dao.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}
	u := domain.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
	return u, nil
}
