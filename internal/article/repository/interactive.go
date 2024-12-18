package repository

import (
	"context"

	"oj/internal/article/repository/cache"
	"oj/internal/article/repository/dao"
)

type InteractiveArtRepository struct {
	dao   *dao.InteractiveDao
	cache *cache.InteractiveCache
}

func NewInteractiveArtRepository(dao *dao.InteractiveDao, cache *cache.InteractiveCache) *InteractiveArtRepository {
	return &InteractiveArtRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *InteractiveArtRepository) IncrReadCnt(ctx context.Context, biz string, bizId uint64) error {
	err := r.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}

	return r.cache.IncrReadCnt(ctx, biz, bizId)
}

func (r *InteractiveArtRepository) IncrLikeCnt(ctx context.Context, biz string, bizId, uid uint64) error {
	// 先插入点赞，然后更新点赞数据，最后更新缓存
	err := r.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	return r.cache.IncrLikeCnt(ctx, biz, bizId)
}

func (r *InteractiveArtRepository) DecrLikeCnt(ctx context.Context, biz string, bizId uint64, uid uint64) error {
	// 先软删除点赞记录，然后更新点赞数据，最后更新缓存
	err := r.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	return r.cache.DecrLikeCnt(ctx, biz, bizId)
}
