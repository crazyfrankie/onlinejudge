package repository

import (
	"context"
	"time"

	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository/dao"
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

func (repo *ArticleRepository) onlineArticleDaoToDomain(art dao.OnlineArticle) domain.Article {
	return domain.Article{
		ID:      art.ID,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorID,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
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

func (repo *ArticleRepository) GetListByID(ctx context.Context, offset, limit int) ([]domain.Article, error) {
	res, err := repo.dao.GetListByID(ctx, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}

	var arts []domain.Article
	for _, art := range res {
		arts = append(arts, repo.articleDaoToDomain(art))
	}

	return arts, nil
}

func (repo *ArticleRepository) GetPubListByID(ctx context.Context, offset, limit int) ([]domain.Article, error) {
	res, err := repo.dao.GetPubListByID(ctx, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}

	var arts []domain.Article
	for _, art := range res {
		arts = append(arts, repo.onlineArticleDaoToDomain(art))
	}

	return arts, nil
}

func (repo *ArticleRepository) GetByID(ctx context.Context, aid uint64) (domain.Article, error) {
	art, err := repo.dao.GetByID(ctx, aid)
	if err != nil {
		return domain.Article{}, err
	}

	return repo.articleDaoToDomain(art), nil
}

func (repo *ArticleRepository) GetPubByID(ctx context.Context, aid uint64) (domain.Article, error) {
	art, err := repo.dao.GetPubByID(ctx, aid)
	if err != nil {
		return domain.Article{}, err
	}

	return repo.onlineArticleDaoToDomain(art), nil
}
