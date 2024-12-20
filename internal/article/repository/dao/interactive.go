package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	
	"oj/internal/article/domain"
)

type InteractiveDao struct {
	db *gorm.DB
}

func NewInteractiveDao(db *gorm.DB) *InteractiveDao {
	return &InteractiveDao{
		db: db,
	}
}

func (dao *InteractiveDao) IncrReadCnt(ctx context.Context, biz string, bizId uint64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		BizID:   bizId,
		Biz:     biz,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao *InteractiveDao) InsertLikeInfo(ctx context.Context, biz string, bizId, uid uint64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(
				map[string]any{
					"utime":  now,
					"status": 1,
				}),
		}).Create(&UserLike{
			Biz:    biz,
			BizID:  bizId,
			UID:    uid,
			Ctime:  now,
			Utime:  now,
			Status: 1,
		}).Error

		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			BizID:   bizId,
			Biz:     biz,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (dao *InteractiveDao) DeleteLikeInfo(ctx context.Context, biz string, bizId uint64, uid uint64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&UserLike{}).Where("uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).Updates(map[string]any{
			"utime":  now,
			"status": 0,
		}).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).
			Model(&Interactive{}).
			Where("biz = ? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt - 1"),
			}).Error
	})
}

func (dao *InteractiveDao) GetInteractiveByID(ctx context.Context, biz string, bizId, uid uint64) (domain.Interactive, error) {
	var inter Interactive
	var userLike UserLike

	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&Interactive{}).Where("biz = ? AND biz_id = ?", biz, bizId).First(&inter).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Model(&UserLike{}).Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&userLike).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Interactive{}, err
	}

	return domain.Interactive{
		Liked:   userLike.Status == 1,
		LikeCnt: inter.LikeCnt,
		ReadCnt: inter.ReadCnt,
	}, nil
}
