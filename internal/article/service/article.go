package service

import (
	"context"
	"oj/internal/article/domain"
)

type ArticleService struct {
	//repo *repository.ArticleRepository
}

func NewArticleService() *ArticleService {
	return &ArticleService{
		//repo: repo,
	}
}

func (svc *ArticleService) Save(ctx context.Context, art domain.Article) (uint64, error) {
	return 1, nil
}
