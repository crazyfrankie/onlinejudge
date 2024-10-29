package dao

import (
	"context"
	"gorm.io/gorm"
	"oj/internal/article/domain"
)

type ArticleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) *ArticleDao {
	return &ArticleDao{
		db: db,
	}
}

func (dao *ArticleDao) Create(ctx context.Context, art domain.Article) (uint64, error) {
	
}
