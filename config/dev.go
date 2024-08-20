//go:build !k8s

package config

// 本地连接

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(localhost:3307)/onlinejudge?charset=utf8mb4&parseTime=true&loc=Local",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
