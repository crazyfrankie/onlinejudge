package ioc

import (
	"github.com/redis/go-redis/v9"
	
	"github.com/crazyfrankie/onlinejudge/infra/contract/token"
	tokenimpl "github.com/crazyfrankie/onlinejudge/infra/impl/token"
)

func InitToken(cmd redis.Cmdable) token.Token {
	return tokenimpl.NewRedisJWTHandler(cmd)
}
