package repository

import (
	"context"
	"oj/user/repository/memory"

	"oj/user/repository/cache"
)

var (
	ErrSendTooMany   = cache.ErrSendTooMany
	ErrVerifyTooMany = cache.ErrVerifyTooMany
)

type CodeRepository struct {
	cache *cache.CodeCache
	mem   *memory.CodeMem
}

func NewCodeRepository(c *cache.CodeCache, mem *memory.CodeMem) *CodeRepository {
	return &CodeRepository{
		cache: c,
		mem:   mem,
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
