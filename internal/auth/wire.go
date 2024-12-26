//go:build wireinject

package auth

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/onlinejudge/internal/auth/jwt"
	ratelimit2 "github.com/crazyfrankie/onlinejudge/internal/auth/ratelimit"
	"github.com/crazyfrankie/onlinejudge/pkg/ratelimit"
)

func InitModule(limiter ratelimit.Limiter, cmd redis.Cmdable) *Module {
	wire.Build(
		jwt.NewRedisJWTHandler,

		ratelimit2.NewBuilder,

		NewLoginJWTMiddlewareBuilder,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
