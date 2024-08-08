package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 创建者模式

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{paths: make([]string, 0)}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(paths string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, paths)
	return l
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 用 Go 的方式进行编码解码 需要先注册才能使用
		gob.Register(time.Now())

		// 不需要登录校验的
		for _, path := range l.paths {
			if c.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(c)
		identifier := sess.Get("identifier")
		if identifier == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Set("identifier", identifier)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})

		now := time.Now().UnixMilli()
		// 第一次登录 还没刷新过
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}

		// update_time 是有的
		updateTimeVal, _ := updateTime.(int64)
		if (now - updateTimeVal) > 60*1000 {
			sess.Set("update_time", now)
			sess.Save()
		}
	}
}
