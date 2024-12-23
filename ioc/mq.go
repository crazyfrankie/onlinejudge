package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"oj/internal/article/event"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}

	client, err := sarama.NewClient([]string{cfg.Addr}, saramaCfg)
	if err != nil {
		panic(err)
	}

	return client
}

func NewConsumers(csm *event.ArticleConsumer) []event.Consumer {
	return []event.Consumer{csm}
}
