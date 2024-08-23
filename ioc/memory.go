package ioc

import (
	"time"

	"github.com/patrickmn/go-cache"
)

func InitGoMem() *cache.Cache {
	return cache.New(time.Minute*10, time.Minute*15)
}
