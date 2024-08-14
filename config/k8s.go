//go:build k8s

package config

// k8s 部署

var Config = config{
	DB: DBConfig{
		DSN: "root:root041126@tcp(oj-mysql:11309)/onlinejudge?charset=utf8mb4&parseTime=true&loc=Local",
	},
	Redis: RedisConfig{
		Addr: "oj-redis:11409",
	},
}
