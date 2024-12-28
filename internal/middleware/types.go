package middleware

import (
	"github.com/crazyfrankie/onlinejudge/internal/middleware/jwt"
)

type Handler = jwt.Handler

type Module struct {
	Hdl Handler
}
