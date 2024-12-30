package ioc

import (
	"github.com/IBM/sarama"

	"github.com/crazyfrankie/onlinejudge/config"
	"github.com/crazyfrankie/onlinejudge/internal/article/event"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{config.GetConf().Kafka.Addr}, saramaCfg)
	if err != nil {
		panic(err)
	}

	return client
}

func NewConsumers(csm event.Consumer) []event.Consumer {
	return []event.Consumer{csm}
}
