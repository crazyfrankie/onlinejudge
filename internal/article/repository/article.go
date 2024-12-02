package repository

import (
	"context"

	"oj/internal/article/domain"
	"oj/internal/article/repository/dao"
)

type ArticleRepository struct {
	dao *dao.GORMArticleDao
}

func NewArticleRepository(dao *dao.GORMArticleDao) *ArticleRepository {
	return &ArticleRepository{
		dao: dao,
	}
}

func (repo *ArticleRepository) CreateDraft(ctx context.Context, art domain.Article) (uint64, error) {
	return repo.dao.CreateDraft(ctx, repo.articleDomainToDao(art))
}

func (repo *ArticleRepository) UpdateDraft(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateDraftByID(ctx, repo.articleDomainToDao(art))
}

func (repo *ArticleRepository) Sync(ctx context.Context, art domain.Article) (uint64, error) {
	return repo.dao.SyncToPublish(ctx, repo.articleDomainToDao(art))
}

func (repo *ArticleRepository) onlineArticleDomainToDao(art domain.Article) dao.OnlineArticle {
	return dao.OnlineArticle{
		ID:       art.ID,
		Title:    art.Title,
		Content:  art.Content,
		AuthorID: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *ArticleRepository) articleDomainToDao(art domain.Article) dao.Article {
	return dao.Article{
		ID:       art.ID,
		Content:  art.Content,
		Title:    art.Title,
		AuthorID: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *ArticleRepository) SyncStatus(ctx context.Context, id uint64, authorId uint64, private domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, id, authorId, private.ToUint8())
}
