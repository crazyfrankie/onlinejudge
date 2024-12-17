package repository

import (
	"context"
	"time"

	"oj/internal/article/domain"
	"oj/internal/article/repository/dao"
)

var (
	ErrRecordNotFound = dao.ErrRecordNotFound
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

func (repo *ArticleRepository) articleDaoToDomain(art dao.Article) domain.Article {
	return domain.Article{
		ID:      art.ID,
		Content: art.Content,
		Title:   art.Title,
		Author: domain.Author{
			Id: art.AuthorID,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (repo *ArticleRepository) SyncStatus(ctx context.Context, id uint64, authorId uint64, private domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, id, authorId, private.ToUint8())
}

func (repo *ArticleRepository) List(ctx context.Context, uid uint64, offset, limit int) ([]domain.Article, error) {
	res, err := repo.dao.List(ctx, uid, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}

	var arts []domain.Article
	for _, art := range res {
		arts = append(arts, repo.articleDaoToDomain(art))
	}

	return arts, nil
}

func (repo *ArticleRepository) GetByID(ctx context.Context, artID uint64) (domain.Article, error) {
	art, err := repo.dao.GetByID(ctx, artID)
	if err != nil {
		return domain.Article{}, err
	}

	return repo.articleDaoToDomain(art), nil
}
