package ioc

import (
	"github.com/gin-gonic/gin"
	"oj/internal/article/event"
)

type App struct {
	Server    *gin.Engine
	Consumers []event.Consumer
}

func InitOJ(server *gin.Engine, consumers []event.Consumer) *App {
	return &App{
		Server:    server,
		Consumers: consumers,
	}
}
