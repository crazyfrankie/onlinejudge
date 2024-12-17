package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
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

func (dao *InteractiveDao) IncrLikeCnt(ctx context.Context, biz string, bizId uint64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
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
}

func (dao *InteractiveDao) IncrCollectCnt(ctx context.Context, biz string, bizId uint64) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"collect_cnt": gorm.Expr("collect_cnt + 1"),
			"utime":       now,
		}),
	}).Create(&Interactive{
		BizID:      bizId,
		Biz:        biz,
		CollectCnt: 1,
		Ctime:      now,
		Utime:      now,
	}).Error
}
