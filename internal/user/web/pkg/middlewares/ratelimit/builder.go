package ratelimit

import (
	_ "embed" // 导入 embed 包，用于在编译时将文件嵌入到 Go 二进制文件中
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Builder struct {
	cmd      redis.Cmdable // Redis 客户端接口，用于执行 Redis 命令
	prefix   string        // 限流键的前缀，通常用于区分不同的限流对象
	interval time.Duration // 限流的时间窗口大小，例如 1 秒钟
	rate     int           // 限流阈值，表示在时间窗口内允许的最大请求数
}

// 使用 go:embed 指令嵌入 Lua 脚本文件 "slide_window.lua"
// 该指令会在编译时将文件内容嵌入到 luaScript 变量中
//
//go:embed slide_window.lua
var luaScript string // 保存从 "slide_window.lua" 文件中嵌入的 Lua 脚本内容

// NewBuilder 创建一个限流器的构建器
// 参数 cmd 表示 Redis 客户端，interval 表示限流时间窗口，rate 表示限流阈值
func NewBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *Builder {
	return &Builder{
		cmd:      cmd,
		prefix:   "ip-limit", // 默认限流前缀为 "ip-limit"
		interval: interval,
		rate:     rate,
	}
}

// Prefix 允许设置自定义前缀，用于区分不同的限流对象
func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build 构建 Gin 的中间件，用于处理限流逻辑
func (b *Builder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		limited, err := b.limit(c)
		if err != nil {
			// 如果执行限流时发生错误，返回 500 错误码
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			// 如果请求被限流，返回 429 Too Many Requests 错误码
			log.Println(err)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		// 如果未限流，继续处理下一个中间件或请求处理程序
		c.Next()
	}
}

// limit 执行限流逻辑，通过 Redis 的 Lua 脚本实现滑动窗口限流
func (b *Builder) limit(c *gin.Context) (bool, error) {
	// 生成限流的键，通常为前缀加上客户端 IP 地址
	key := fmt.Sprintf("%s:%s", b.prefix, c.ClientIP())

	// 使用 Redis 的 Eval 方法执行 Lua 脚本
	// 参数为 Lua 脚本内容、键列表和其他参数（窗口大小、限流阈值、当前时间）
	return b.cmd.Eval(c, luaScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
