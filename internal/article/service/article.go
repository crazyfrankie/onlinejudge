package service

import (
	"context"

	"go.uber.org/zap"

	"oj/internal/article/domain"
	"oj/internal/article/repository"
)

type ArticleService interface {
	SaveDraft(ctx context.Context, art domain.Article) (uint64, error)
	Publish(ctx context.Context, art domain.Article) (uint64, error)
}

type articleService struct {
	repo   *repository.ArticleRepository
	logger *zap.Logger
}

func NewArticleService(repo *repository.ArticleRepository, logger *zap.Logger) ArticleService {
	return &articleService{
		repo:   repo,
		logger: logger,
	}
}

// SaveDraft 接口只改制作库
// 如果 art 的 ID 大于0,说明是更新草稿,直接去制作库更新代表保存
// 如果 art 的 ID 小于等于0,说明是新建的草稿,创建草稿到制作库即代表保存
func (svc *articleService) SaveDraft(ctx context.Context, art domain.Article) (uint64, error) {
	if art.ID > 0 {
		err := svc.repo.UpdateDraft(ctx, art)
		return art.ID, err
	}
	return svc.repo.CreateDraft(ctx, art)
}

// Publish 接口先改制作库，然后同步到线上库
// 如果 art 的 ID 大于0,说明是更新，先修改制作库，然后同步到线上库作为发表
// 如果 art 的 ID 小于等于0,说明是新建，先去制作库创建，然后同步到线上库作为发表
func (svc *articleService) Publish(ctx context.Context, art domain.Article) (uint64, error) {
	//var (
	//	id  = art.ID
	//	err error
	//)
	//if art.ID > 0 {
	//	err = svc.repo.Update(ctx, art)
	//} else {
	//	id, err = svc.repo.Create(ctx, art)
	//}
	//if err != nil {
	//	return 0, err
	//}
	//
	//art.ID = id
	//
	//// 重试机制
	//maxRetries := 3
	//for i := 0; i < maxRetries; i++ {
	//	time.Sleep(time.Second * time.Duration(i))
	//	id, err = svc.repo.Sync(ctx, art)
	//	if err == nil {
	//		break
	//	}
	//	svc.logger.Error("部分失败:保存到线上库失败", zap.Uint64("art_id", art.ID), zap.Error(err))
	//}
	//
	//// 所有重试失败，记录日志
	//if err != nil {
	//	svc.logger.Error("部分失败:重试保存到线上库彻底失败", zap.Uint64("artID", art.ID), zap.Error(err))
	//	// 接入告警系统,手动处理一下
	//	// 走异步,直接保存到本地文件
	//	// 走 Canal,那上面的操作都不需要
	//	// 打 MQ
	//}
	//
	//return id, err
	return svc.repo.Sync(ctx, art)
}
