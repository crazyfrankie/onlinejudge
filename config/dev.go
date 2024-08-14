//go:build !k8s

package config

// 本地连接

var Config = config{
	DB: DBConfig{
		DSN: "localhost:3306",
	},
	Redis: RedisConfig{
		Addr: "172.21.21.73:8838",
	},
}
