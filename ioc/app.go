package ioc

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/onlinejudge/internal/article/event"
)

type App struct {
	Server    *gin.Engine
	Consumers []event.Consumer
}
