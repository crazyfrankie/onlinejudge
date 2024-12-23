package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()

	// 正常来说，一个消费者都是属于一个消费者组的
	// 一个消费者组就是一个业务
	consumerGroup, err := sarama.NewConsumerGroup(addr, "test_group", cfg)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = consumerGroup.Consume(ctx, []string{"test_topic"}, testConsumerGroupHandler{})
	// 消费结束就会到这里,阻塞调用
	t.Log(err)
}

type testConsumerGroupHandler struct{}

func (l testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// 指定偏移量消费
	partitions := session.Claims()["test_topic"]
	for _, p := range partitions {
		session.ResetOffset("test_topic", p, sarama.OffsetOldest, "")
		// 或者根据业务指定偏移量
		session.ResetOffset("test_topic", p, 123, "")
	}

	return nil
}

func (l testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

// 异步消费
func (l testConsumerGroupHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage

		for i := 0; i < batchSize; i++ {
			done := false
			select {
			case <-ctx.Done():
				// 代表超时了
				done = true
			case msg, ok := <-msgs:
				// 消费者被关闭了
				if !ok {
					cancel()
					return nil
				}
				last = msg
				eg.Go(func() error {
					time.Sleep(time.Second)
					// 可以在这里重试
					log.Println(string(msg.Value))
					return nil
				})
			}
			if done {
				break
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			// 1. 在这里重试
			// 2. 记录日志
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}

// 同步消费
func (l testConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()

	for msg := range msgs {
		//var bizMsg BizMsg
		//err := json.Unmarshal(msg.Value, &bizMsg)
		//if err != nil {
		//
		//}
		log.Println(string(msg.Value))
		// 标记为消费成功
		session.MarkMessage(msg, "")
	}

	return nil
}

type BizMsg struct {
	Name string
}
