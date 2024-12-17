package dao

import (
	"bytes"
	"context"
	"gorm.io/gorm/clause"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
)

type OSSArticleDao struct {
	oss *s3.S3
	GORMArticleDao
	bucket *string
}

func NewOSSArticleDao(oss *s3.S3, db *gorm.DB) *OSSArticleDao {
	return &OSSArticleDao{
		oss: oss,
		GORMArticleDao: GORMArticleDao{
			db: db,
		},
		bucket: nil,
	}
}

func (o *OSSArticleDao) Sync(ctx context.Context, art Article) (uint64, error) {
	// 保存制作库
	// 保存线上库
	// 把 Content 上传到 OSS
	var (
		id = art.ID
	)
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var err error

		txDao := NewArticleDao(tx)
		if id > 0 {
			err = txDao.UpdateDraftByID(ctx, art)
		} else {
			id, err = txDao.CreateDraft(ctx, art)
		}
		if err != nil {
			return err
		}

		art.ID = id
		publishArt := OnlineArticle(art)
		now := time.Now().UnixMilli()
		publishArt.Ctime = now
		publishArt.Utime = now

		// 线上库不保存 Content, 要准备上传到 OSS
		publishArt.Content = ""
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			// Columns 哪些列冲突
			Columns: []clause.Column{{Name: "id"}},
			// 如果是更新，则更新以下字段
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  art.Title,
				"utime":  now,
				"status": art.Status,
			}),
			// DoNothing: 数据冲突了啥也不干
			// Where: 数据冲突了，并且符合 WHERE 条件的就会执行更新
		}).Create(&art).Error
	})
	if err != nil {
		return 0, err
	}

	// 保存到 OSS
	// 需要有监控，重试，补偿机制
	_, err = o.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      o.bucket,
		Key:         ToPtr(strconv.FormatInt(int64(id), 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ToPtr("text/plain;charset=utf-8"),
	})

	return id, err
}

//func (o *OSSArticleDao) SyncStatus(ctx context.Context, id uint64, authorId uint64, status uint8) error {
//	now := time.Now().UnixMilli()
//	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//		res := tx.Model(&Article{}).
//			Where("id = ? AND author_id = ?", id, authorId).
//			Updates(map[string]any{
//				"status": status,
//				"utime":  now,
//			})
//
//		if res.Error != nil {
//			// 数据库有问题
//			return res.Error
//		}
//		if res.RowsAffected != 1 {
//			return fmt.Errorf("有人误操作,uid :%d", id)
//		}
//
//		if status == domain.ArticleStatusPrivate.ToUint8() {
//			o.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
//				Bucket: o.bucket,
//				Key:    ToPtr(strconv.FormatInt(int64(id), 10)),
//			})
//		}
//
//		return tx.Model(&OnlineArticle{}).
//			Where("id = ? ", id).
//			Updates(map[string]any{
//				"status": status,
//				"utime":  now,
//			}).Error
//	})
//}

func ToPtr(s string) *string {
	return &s
}
