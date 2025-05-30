package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/kr/pretty"
	"github.com/spf13/viper"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env    string
	MySQL  MySQL  `yaml:"mysql"`
	Redis  Redis  `yaml:"redis"`
	WeChat WeChat `yaml:"wechat"`
	Kafka  Kafka  `yaml:"kafka"`
	Judge  Judge  `yaml:"judge"`
}

type MySQL struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	ConnMaxLifeTime int    `yaml:"connMaxLifeTime"`
}

type Redis struct {
	Address      string `yaml:"address"`
	PoolSize     int    `yaml:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns"`
}

type WeChat struct {
	AppId  string `yaml:"appId"`
	AppKey string `yaml:"appKey"`
}

type Kafka struct {
	Addr string `yaml:"addr"`
}

type Judge struct {
	Addr string `yaml:"addr"`
}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {
	prefix := "config"

	err := godotenv.Load(filepath.Join(prefix, ".env"))
	if err != nil {
		panic(err)
	}

	contentFilePath := filepath.Join(prefix, filepath.Join(GetEnv(), "conf.yaml"))
	viper.SetConfigFile(contentFilePath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	conf = new(Config)
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	conf.Env = GetEnv()
	// 打印配置，方便调试
	pretty.Printf("%+v\n", conf)
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}
