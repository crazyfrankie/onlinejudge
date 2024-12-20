// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"oj/internal/article"
	"oj/internal/judgement"
	"oj/internal/problem"
	"oj/internal/user"
	"oj/internal/user/middleware/jwt"
)

// Injectors from wire.go:

func InitGin() *gin.Engine {
	cmdable := InitRedis()
	limiter := InitSlideWindow(cmdable)
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := GinMiddlewares(limiter, handler)
	db := InitDB()
	userHandler := user.InitUserHandler(cmdable, db)
	problemHandler := problem.InitProblemHandler(cmdable, db)
	oAuthWeChatHandler := user.InitOAuthWeChatHandler(cmdable, db)
	localSubmitHandler := judgement.InitLocalJudgement(cmdable, db)
	submissionHandler := judgement.InitRemoteJudgement(cmdable, db)
	oAuthGithubHandler := user.InitOAuthGithubHandler(cmdable, db)
	articleHandler := article.InitArticleHandler(db)
	engine := InitWebServer(v, userHandler, problemHandler, oAuthWeChatHandler, localSubmitHandler, submissionHandler, oAuthGithubHandler, articleHandler)
	return engine
}

// wire.go:

var BaseSet = wire.NewSet(InitDB, InitRedis)
