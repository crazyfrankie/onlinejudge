package dao

import (
	"context"
	"database/sql"
	"errors"
	"gorm.io/gorm"
	"time"

	"github.com/crazyfrankie/onlinejudge/internal/user/domain"
)

var (
	ErrUserDuplicateWechat = errors.New("duplicate wechat")
	ErrUserDuplicateGithub = errors.New("duplicate github")
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
	Ctime         int64
	Uptime        int64
}

type UserDao interface {
	Insert(ctx context.Context, user domain.User) error
	InsertByWeChat(ctx context.Context, user domain.User) error
	InsertByGithub(ctx context.Context, user domain.User) error
	FindByName(ctx context.Context, name string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id uint64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	FindByGithub(ctx context.Context, gitId string) (domain.User, error)
	UpdatePassword(ctx context.Context, uid uint64, password string) (domain.User, error)
	UpdateInfo(ctx context.Context, uid uint64, updates map[string]any) error
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{
		db: db,
	}
}

func (dao *GormUserDao) Insert(ctx context.Context, user domain.User) error {
	u := User{
		Phone: user.Phone,
		Name:  user.Name,
	}
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	if err := dao.db.WithContext(ctx).Create(&u).Error; err != nil {
		return err
	}

	return nil
}

func (dao *GormUserDao) InsertByWeChat(ctx context.Context, user domain.User) error {
	u := User{
		WechatOpenID: sql.NullString{
			String: user.WeChatInfo.OpenID,
			Valid:  user.WeChatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: user.WeChatInfo.UnionID,
			Valid:  user.WeChatInfo.UnionID != "",
		},
		Name: user.Name,
	}
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	if err := dao.db.WithContext(ctx).Create(&u).Error; err != nil {
		return err
	}

	return nil
}

func (dao *GormUserDao) InsertByGithub(ctx context.Context, user domain.User) error {
	u := User{
		GithubID: sql.NullString{
			String: user.GithubID,
			Valid:  user.GithubID != "",
		},
		Name: user.Name,
	}
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Uptime = now
	if err := dao.db.WithContext(ctx).Create(&u).Error; err != nil {
		return err
	}

	return nil
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

	return domain.User{
		Id:       user.Id,
		GithubID: user.GithubID.String,
		Name:     user.Name,
	}, nil
}

func (dao *GormUserDao) UpdatePassword(ctx context.Context, uid uint64, password string) (domain.User, error) {
	// 直接更新用户信息
	result := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).Updates(User{
		Password: password,
	})
	if result.Error != nil {
		return domain.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return domain.User{}, errors.New("user not found or no updates made")
	}

	return dao.FindById(ctx, uid)
}

func (dao *GormUserDao) UpdateInfo(ctx context.Context, uid uint64, updates map[string]any) error {
	err := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).Updates(updates).Error
	if err != nil {
		return err
	}

	return nil
}
