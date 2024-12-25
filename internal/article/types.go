package article

import (
	"github.com/crazyfrankie/onlinejudge/internal/article/event"
	"github.com/crazyfrankie/onlinejudge/internal/article/web"
)

type Handler = web.ArticleHandler
type AdminHandler = web.AdminHandler
type Consumer = event.Consumer

type Module struct {
	Hdl      *Handler
	AdminHdl *AdminHandler
	Consumer Consumer
}
