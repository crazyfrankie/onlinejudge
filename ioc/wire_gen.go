// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"oj/user/repository"
	"oj/user/repository/cache"
	"oj/user/repository/dao"
	"oj/user/service"
	"oj/user/web"
)

// Injectors from wire.go:

func InitGin() *gin.Engine {
	cmdable := InitRedis()
	v := GinMiddlewares(cmdable)
	db := InitDB()
	userDao := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := InitWebServer(v, userHandler)
	return engine
}