package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
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
	Email    sql.NullString `gorm:"unique"`
	Phone    string         `gorm:"unique"`
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

func (dao *UserDao) generateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomNumber := rand.Intn(1000000)

	return fmt.Sprintf("%6d", randomNumber)
}

func (dao *UserDao) Insert(ctx context.Context, u *User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	err := dao.db.WithContext(ctx).Create(u).Error
	return handleDBError(err)
}

func (dao *UserDao) FindByName(ctx context.Context, name string) (domain.User, error) {
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

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (domain.User, error) {
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
		Email: user.Email.String,
		Phone: user.Phone,
		Role:  user.Role,
	}
	return u, nil
}

func (dao *UserDao) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user)

	if result.Error != nil {
		// 判断是不是因为没有创建才没找到
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			code := dao.generateCode()
			newUser := User{
				Name:  "用户" + code,
				Phone: phone,
				Email: sql.NullString{
					String: "",
					Valid:  false,
				},
			}
			err := dao.Insert(ctx, &newUser)
			if err != nil {
				// 处理插入时的错误，例如数据库连接问题
				return domain.User{}, err
			}
			u := domain.User{
				Id:    newUser.Id,
				Name:  newUser.Name,
				Phone: newUser.Phone,
				Email: newUser.Email.String,
				Role:  0,
			}
			return u, nil
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
