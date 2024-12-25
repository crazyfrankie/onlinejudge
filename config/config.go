package config

import (
	"github.com/kr/pretty"
	"github.com/spf13/viper"
	"log"
	"os"
	"sync"
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
	//Registry Registry `yaml:"registry"`
}

type MySQL struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	ConnMaxLifeTime int    `yaml:"connMaxLifeTime"`
}

type Redis struct {
	Address      string `yaml:"addr"`
	PoolSize     int    `yaml:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns"`
	//Username string `yaml:"username"`
	//Password string `yaml:"password"`
	//DB       int    `yaml:"db"`
}

type WeChat struct {
	AppId  string `yaml:"appId"`
	AppKey string `yaml:"appKey"`
}

type Kafka struct {
	Addr string `yaml:"addr"`
}

//type Registry struct {
//	RegistryAddress []string `yaml:"registry_address"`
//	Username        string   `yaml:"username"`
//	Password        string   `yaml:"password"`
//}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {
	viper.SetConfigFile("config/dev.yaml")
	if err := viper.ReadInConfig(); err != nil {
		// 这里可以改成更优雅的错误处理
		log.Fatalf("Failed to read config file: %v", err)
	}

	conf = new(Config)
	if err := viper.Unmarshal(&conf); err != nil {
		// 更优雅的错误处理
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	conf.Env = GetEnv() // Ensure GetEnv() works as expected
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
