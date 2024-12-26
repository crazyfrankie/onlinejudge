package auth

import (
	"github.com/crazyfrankie/onlinejudge/internal/auth/jwt"
	"github.com/crazyfrankie/onlinejudge/internal/auth/ratelimit"
)

type Handler = jwt.Handler
type Builder = ratelimit.Builder
type JWTBuilder = LoginJWTMiddlewareBuilder
type Module struct {
	Hdl        Handler
	Builder    *Builder
	JWTBuilder *JWTBuilder
}
