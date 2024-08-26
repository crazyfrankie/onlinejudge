//go:build !k8s

package config

// 本地连接

var Config = config{
	DB: DBConfig{
		DSN:             "root:123456@tcp(localhost:3306)/onlinejudge?charset=utf8mb4&parseTime=true&loc=Local",
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 60,
	},
	Redis: RedisConfig{
		Addr:         "localhost:8838",
		PoolSize:     15,
		MinIdleConns: 5,
	},
}
