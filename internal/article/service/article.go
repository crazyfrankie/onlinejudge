package service

import (
	"context"
	"errors"
	"log"
	"strconv"

	"go.uber.org/zap"

	"github.com/crazyfrankie/onlinejudge/common/constant"
	er "github.com/crazyfrankie/onlinejudge/common/errors"
	"github.com/crazyfrankie/onlinejudge/internal/article/domain"
	"github.com/crazyfrankie/onlinejudge/internal/article/event"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository"
)

type ArticleService interface {
	// B 端接口
	SaveDraft(ctx context.Context, art domain.Article) (uint64, error)
	Publish(ctx context.Context, art domain.Article) (uint64, error)
	WithDraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, offset, limit int) ([]domain.Article, error)
	Detail(ctx context.Context, uid uint64, artID string) (domain.Article, error)

	// C 端接口
	PubDetail(ctx context.Context, uid uint64, artID string) (domain.Article, error)
	PubList(ctx context.Context, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     *repository.ArticleRepository
	logger   *zap.Logger
	producer event.Producer
}

func NewArticleService(repo *repository.ArticleRepository, logger *zap.Logger, producer event.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		logger:   logger,
		producer: producer,
	}
}

// SaveDraft 接口只改制作库
// 如果 art 的 ID 大于0,说明是更新草稿,直接去制作库更新代表保存
// 如果 art 的 ID 小于等于0,说明是新建的草稿,创建草稿到制作库即代表保存
func (svc *articleService) SaveDraft(ctx context.Context, art domain.Article) (uint64, error) {
	art.Status = domain.ArticleStatusUnPublished
	if art.ID > 0 {
		err := svc.repo.UpdateDraft(ctx, art)
		if err != nil {
			return 0, er.NewBusinessError(constant.ErrUpdateDraft)
		}

		return art.ID, nil
	}

	id, err := svc.repo.CreateDraft(ctx, art)
	if err != nil {
		return 0, er.NewBusinessError(constant.ErrAddDraft)
	}

	return id, nil
}

// Publish 接口先改制作库，然后同步到线上库
// 如果 art 的 ID 大于0,说明是更新，先修改制作库，然后同步到线上库作为发表
// 如果 art 的 ID 小于等于0,说明是新建，先去制作库创建，然后同步到线上库作为发表
func (svc *articleService) Publish(ctx context.Context, art domain.Article) (uint64, error) {
	id, err := svc.repo.Sync(ctx, art)
	if err != nil {
		return 0, er.NewBusinessError(constant.ErrSyncPublish)
	}

	return id, nil
}

func (svc *articleService) WithDraw(ctx context.Context, art domain.Article) error {
	err := svc.repo.SyncStatus(ctx, art.ID, art.Author.Id, domain.ArticleStatusPrivate)
	if err != nil {
		return er.NewBusinessError(constant.ErrWithdrawArt)
	}

	return nil
}

func (svc *articleService) List(ctx context.Context, offset, limit int) ([]domain.Article, error) {
	res, err := svc.repo.GetListByID(ctx, offset, limit)
	if err != nil {
		return []domain.Article{}, er.NewBusinessError(constant.ErrInternalServer)
	}

	return res, nil
}

func (svc *articleService) PubList(ctx context.Context, offset, limit int) ([]domain.Article, error) {
	res, err := svc.repo.GetPubListByID(ctx, offset, limit)
	if err != nil {
		return []domain.Article{}, er.NewBusinessError(constant.ErrInternalServer)
	}

	return res, nil
}

func (svc *articleService) Detail(ctx context.Context, uid uint64, artID string) (domain.Article, error) {
	id, _ := strconv.Atoi(artID)

	art, err := svc.repo.GetByID(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return domain.Article{}, er.NewBusinessError(constant.ErrArticleNotFound)
		}

		return domain.Article{}, er.NewBusinessError(constant.ErrInternalServer)
	}

	if uid != art.Author.Id {
		return domain.Article{}, er.NewBusinessError(constant.ErrForbidden)
	}

	return art, nil
}

func (svc *articleService) PubDetail(ctx context.Context, uid uint64, artID string) (domain.Article, error) {
	id, _ := strconv.Atoi(artID)

	art, err := svc.repo.GetPubByID(ctx, uint64(id))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return domain.Article{}, er.NewBusinessError(constant.ErrArticleNotFound)
		}

		return domain.Article{}, er.NewBusinessError(constant.ErrInternalServer)
	}

	go func() {
		er := svc.producer.ProduceReadEvent(ctx, event.ReadEvent{Aid: uint64(id), Uid: uid})
		if er != nil {
			log.Printf("增加阅读计数失败:%s:aid:%d:uid:%d", er.Error(), id, uid)
		}
	}()

	return art, nil
}
