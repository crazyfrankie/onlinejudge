package config

type config struct {
	DB    DBConfig
	Redis RedisConfig
}

type DBConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Addr         string
	PoolSize     int
	MinIdleConns int
}
