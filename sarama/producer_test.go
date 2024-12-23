package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"testing"
)

var addr = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	// 同步发送
	//cfg := sarama.NewConfig()
	//cfg.Producer.Return.Successes = true
	//producer, err := sarama.NewSyncProducer(addr, cfg)
	//assert.NoError(t, err)
	//
	//_, _, err = producer.SendMessage(&sarama.ProducerMessage{
	//	Topic: "test_topic",
	//	Value: sarama.StringEncoder("this is a test message A"),
	//})
	//assert.NoError(t, err)

	// 异步发送
	cfg := sarama.NewConfig()

	// 为了后面拿到结果,根据需求设置
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	require.NoError(t, err)

	// 异步发送是一个 Channel
	msgCh := producer.Input()
	// 初始化 Channel, 并发送
	msgCh <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("this is a test message B"),
	}

	// 处理结果
	errCh := producer.Errors()
	succCh := producer.Successes()
	select {
	case err := <-errCh:
		t.Log("发送出问题了", err.Error())
	case <-succCh:
		t.Log("发送成功了")
	}
}
