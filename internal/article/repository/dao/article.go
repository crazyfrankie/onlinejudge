package dao

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ArticleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) *ArticleDao {
	return &ArticleDao{
		db: db,
	}
}

func (dao *ArticleDao) Create(ctx context.Context, art Article) (uint64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.ID, err
}

func (dao *ArticleDao) UpdateByID(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now

	updates := make(map[string]interface{})

	// 只添加非空值到映射
	if art.Title != "" {
		updates["title"] = art.Title
	}
	if art.Content != "" {
		updates["content"] = art.Content
	}
	updates["utime"] = art.Utime

	result := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", art.ID, art.AuthorID).Updates(updates)

	if result.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id :%d, author_id: %d", art.ID, art.AuthorID)
	}

	return result.Error
}
