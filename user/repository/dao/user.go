package dao

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"oj/user/domain"
)

type User struct {
	Id       uint64 `gorm:"primaryKey,autoIncrement"`
	Name     string `gorm:"unique;not null"`
	Password string
	Email    sql.NullString `gorm:"unique"`
	Phone    string         `gorm:"unique"`
	Role     uint8
	AboutMe  string
	Birthday sql.NullTime
	Ctime    int64
	Uptime   int64
}

type UserDao interface {
	Insert(ctx context.Context, u *User) error
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{
		db: db,
	}
}

var (
	ErrUserDuplicateEmail = errors.New("duplicate email")
	ErrUserDuplicateName  = errors.New("duplicate name")
	ErrUserDuplicatePhone = errors.New("duplicate phone")
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
		} else if strings.Contains(mysqlErr.Message, "phone") {
			return ErrUserDuplicatePhone
		}
	}
	return err
}

func (dao *GormUserDao) Insert(ctx context.Context, u *User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	err := dao.db.WithContext(ctx).Create(u).Error
	return handleDBError(err)
}

func (dao *GormUserDao) FindByName(ctx context.Context, name string) (domain.User, error) {
	var user User

	result := dao.db.WithContext(ctx).Where("name = ?", name).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}
	u := domain.User{
		Id:       user.Id,
		Password: user.Password,
		Role:     user.Role,
	}
	return u, nil
}

func (dao *GormUserDao) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var user User

	result := dao.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}
	u := domain.User{
		Id:       user.Id,
		Password: user.Password,
		Role:     user.Role,
	}
	return u, nil
}

func (dao *GormUserDao) FindById(ctx context.Context, id uint64) (domain.User, error) {
	var user User

	result := dao.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	u := domain.User{
		Id:       user.Id,
		Name:     user.Name,
		Email:    user.Email.String,
		Phone:    user.Phone,
		Birthday: user.Birthday.Time,
		AboutMe:  user.AboutMe,
		Role:     user.Role,
	}
	return u, nil
}

func (dao *GormUserDao) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user)

	if result.Error != nil {
		// 判断是不是因为没有创建才没找到
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, result.Error
	}

	newUser := domain.User{
		Id:       user.Id,
		Name:     user.Name,
		Password: user.Password,
		Email:    user.Email.String,
		Phone:    user.Phone,
		Role:     user.Role,
	}
	return newUser, nil
}
