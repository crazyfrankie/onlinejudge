package dao

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) *ArticleDao {
	return &ArticleDao{
		db: db,
	}
}

func (dao *ArticleDao) CreateDraft(ctx context.Context, art Article) (uint64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.ID, err
}

func (dao *ArticleDao) UpdateDraftByID(ctx context.Context, art Article) error {
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

func (dao *ArticleDao) SyncToPublish(ctx context.Context, art Article) (uint64, error) {
	var (
		id = art.ID
	)
	// 事务的闭包
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error

		// 采用这个事务去创建连接
		txDao := NewArticleDao(tx)
		if id > 0 {
			err = txDao.UpdateDraftByID(ctx, art)
		} else {
			id, err = txDao.CreateDraft(ctx, art)
		}
		if err != nil {
			return err
		}

		return txDao.Upsert(ctx, OnlineArticle{
			ID:       art.ID,
			Title:    art.Title,
			Content:  art.Content,
			AuthorID: art.AuthorID,
		})
	})
	return id, err
}

func (dao *ArticleDao) Upsert(ctx context.Context, art OnlineArticle) error {
	// 实现 INSERT OR UPDATE
	// SQL:
	// INSERT xxx ON DUPLICATE KEY UPDATE xxx(如果是更新,xxx 代表要更新的列)

	// GORM 实现:
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		// Columns 哪些列冲突
		Columns: []clause.Column{{Name: "id"}},
		// 如果是更新，则更新以下字段
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		}),
		// DoNothing: 数据冲突了啥也不干
		// Where: 数据冲突了，并且符合 WHERE 条件的就会执行更新
	}).Create(&art).Error
}

func (dao *ArticleDao) SyncStatus(ctx context.Context, id uint64, authorId uint64, private uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, authorId).
			Updates(map[string]any{
				"status": private,
				"utime":  now,
			})

		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("有人误操作,uid :%d", id)
		}

		return tx.Model(&Article{}).
			Where("id = ? ", id).
			Updates(map[string]any{
				"status": private,
				"utime":  now,
			}).Error
	})
}
