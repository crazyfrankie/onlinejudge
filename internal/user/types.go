package user

import (
	"github.com/crazyfrankie/onlinejudge/internal/user/middleware/jwt"
	"github.com/crazyfrankie/onlinejudge/internal/user/web"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
)

type Handler = web.UserHandler
type JWTHdl = jwt.Handler
type GithubHandler = third.OAuthGithubHandler
type WeChatHandler = third.OAuthWeChatHandler
type Module struct {
	Hdl       *Handler
	JWTHdl    JWTHdl
	GithubHdl *GithubHandler
	WeChatHdl *WeChatHandler
}
