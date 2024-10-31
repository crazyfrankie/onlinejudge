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
	return repo.dao.Create(ctx, repo.domainToDao(art))
}

func (repo *ArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateByID(ctx, repo.domainToDao(art))
}

func (repo *ArticleRepository) domainToDao(art domain.Article) dao.Article {
	return dao.Article{
		ID:       art.ID,
		Content:  art.Content,
		Title:    art.Title,
		AuthorID: art.Author.Id,
	}
}

func (repo *ArticleRepository) daoToDomain(art dao.Article) domain.Article {
	return domain.Article{
		Content: art.Content,
		Title:   art.Title,
		Author: domain.Author{
			Id: art.AuthorID,
		},
	}
}
