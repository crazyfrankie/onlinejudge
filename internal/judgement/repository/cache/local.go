package cache

import "github.com/redis/go-redis/v9"

type LocalSubmitCache interface {
}

type LocalSubmissionCache struct {
	cmd redis.Cmdable
}

func NewLocalSubmitCache(cmd redis.Cmdable) LocalSubmitCache {
	return &LocalSubmissionCache{
		cmd: cmd,
	}
}
