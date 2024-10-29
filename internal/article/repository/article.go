package repository

import (
	"context"
	"oj/internal/article/domain"

	"oj/internal/article/repository/dao"
)

type ArticleRepository struct {
	dao *dao.ArticleDao
}

func NewArticleRepository(dao *dao.ArticleDao) *ArticleRepository {
	return &ArticleRepository{
		dao: dao,
	}
}

func (repo *ArticleRepository) Create(ctx context.Context, art domain.Article) (uint64, error) {
	return 1, nil
}
