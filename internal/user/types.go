package user

import (
	"github.com/crazyfrankie/onlinejudge/internal/user/web"
	"github.com/crazyfrankie/onlinejudge/internal/user/web/third"
)

type Handler = web.UserHandler
type GithubHandler = third.OAuthGithubHandler
type WeChatHandler = third.OAuthWeChatHandler

type Module struct {
	Hdl       *Handler
	GithubHdl *GithubHandler
	WeChatHdl *WeChatHandler
}
