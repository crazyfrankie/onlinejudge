package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type GORMArticleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) *GORMArticleDao {
	return &GORMArticleDao{
		db: db,
	}
}

func (dao *GORMArticleDao) CreateDraft(ctx context.Context, art Article) (uint64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.ID, err
}

func (dao *GORMArticleDao) UpdateDraftByID(ctx context.Context, art Article) error {
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

func (dao *GORMArticleDao) SyncToPublish(ctx context.Context, art Article) (uint64, error) {
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

func (dao *GORMArticleDao) Upsert(ctx context.Context, art OnlineArticle) error {
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
			"status":  art.Status,
		}),
		// DoNothing: 数据冲突了啥也不干
		// Where: 数据冲突了，并且符合 WHERE 条件的就会执行更新
	}).Create(&art).Error
}

func (dao *GORMArticleDao) SyncStatus(ctx context.Context, id uint64, authorId uint64, private uint8) error {
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

		return tx.Model(&OnlineArticle{}).
			Where("id = ? ", id).
			Updates(map[string]any{
				"status": private,
				"utime":  now,
			}).Error
	})
}

func (dao *GORMArticleDao) GetListByID(ctx context.Context, offset, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&arts).Error

	return arts, err
}

func (dao *GORMArticleDao) GetPubListByID(ctx context.Context, offset int, limit int) ([]OnlineArticle, error) {
	var arts []OnlineArticle
	err := dao.db.WithContext(ctx).Model(&OnlineArticle{}).
		Order("utime DESC").
		Offset(offset).
		Limit(limit).
		Find(&arts).Error

	return arts, err
}

func (dao *GORMArticleDao) GetByID(ctx context.Context, id uint64) (Article, error) {
	var art Article

	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Article{}, ErrRecordNotFound
		}

		return Article{}, err
	}

	return art, nil
}

func (dao *GORMArticleDao) GetPubByID(ctx context.Context, aid uint64) (OnlineArticle, error) {
	var art OnlineArticle

	err := dao.db.WithContext(ctx).Model(&OnlineArticle{}).Where("id = ?", aid).First(&art).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OnlineArticle{}, ErrRecordNotFound
		}

		return OnlineArticle{}, err
	}

	return art, nil
}
