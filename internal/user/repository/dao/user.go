package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"

	"oj/internal/user/domain"
)

var (
	ErrUserDuplicateEmail  = errors.New("duplicate email")
	ErrUserDuplicateName   = errors.New("duplicate name")
	ErrUserDuplicatePhone  = errors.New("duplicate phone")
	ErrUserDuplicateWechat = errors.New("duplicate wechat")
	ErrUserNotFound        = errors.New("user not found")
	ErrRecordNotFound      = errors.New("record not found")
)

type User struct {
	Id            uint64 `gorm:"primaryKey,autoIncrement"`
	Password      string
	Name          string         `gorm:"unique;not null"`
	Phone         string         `gorm:"unique"`
	GithubID      sql.NullString `gorm:"unique"`
	Email         sql.NullString `gorm:"unique"`
	WechatUnionID sql.NullString
	WechatOpenID  sql.NullString `gorm:"unique"`
	Birthday      sql.NullTime
	Role          uint8
	Ctime         int64
	Uptime        int64
}

type UserDao interface {
	Insert(ctx context.Context, user domain.User) error
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	FindByGithub(ctx context.Context, gitId string) (domain.User, error)
	UpdatePassword(ctx context.Context, user domain.User) error
	UpdateName(ctx context.Context, user domain.User) error
	UpdateBirthday(ctx context.Context, user domain.User) error
	UpdateEmail(ctx context.Context, user domain.User) error
	UpdateRole(ctx context.Context, user domain.User) error
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{
		db: db,
	}
}

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

func (dao *GormUserDao) Insert(ctx context.Context, user domain.User) error {
	// 开启事务
	tx := dao.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var existingUser User
	if err := tx.WithContext(ctx).Where("phone = ?", user.Phone).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback() // 查询出错则回滚事务
			return err
		}
	} else {
		tx.Rollback() // 如果用户已存在则回滚事务
		return ErrUserDuplicatePhone
	}

	u := User{
		Phone: user.Phone,
		Role:  0,
		Name:  user.Name,
	}
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	if err := tx.WithContext(ctx).Create(&u).Error; err != nil {
		tx.Rollback() // 插入失败则回滚事务
		return handleDBError(err)
	}

	return tx.Commit().Error
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

func (dao *GormUserDao) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&user)

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
		WeChatInfo: domain.WeChatInfo{
			OpenID:  user.WechatOpenID.String,
			UnionID: user.WechatUnionID.String,
		},
		Role: user.Role,
	}
	return newUser, nil
}

func (dao *GormUserDao) FindByGithub(ctx context.Context, gitId string) (domain.User, error) {
	var user User

	result := dao.db.WithContext(ctx).Where("github_id = ?", gitId).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrRecordNotFound
		}
		return domain.User{}, result.Error
	}

	return domain.User{}, nil
}

func (dao *GormUserDao) UpdatePassword(ctx context.Context, user domain.User) error {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(User{
		Password: user.Password,
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no updates made")
	}

	return nil
}

func (dao *GormUserDao) UpdateName(ctx context.Context, user domain.User) error {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(User{
		Name: user.Name,
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no updates made")
	}

	return nil
}

func (dao *GormUserDao) UpdateBirthday(ctx context.Context, user domain.User) error {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(User{
		Birthday: sql.NullTime{
			Time:  user.Birthday,
			Valid: true,
		},
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no updates made")
	}

	return nil
}

func (dao *GormUserDao) UpdateEmail(ctx context.Context, user domain.User) error {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(User{
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no updates made")
	}

	return nil
}

func (dao *GormUserDao) UpdateRole(ctx context.Context, user domain.User) error {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(User{
		Role: user.Role,
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no updates made")
	}

	return nil
}
