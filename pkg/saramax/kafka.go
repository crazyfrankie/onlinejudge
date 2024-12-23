package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Handler[T any] struct {
	l  *zap.Logger
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l *zap.Logger, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化消息失败",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset))
			continue
		}

		err = h.fn(msg, t)
		// 可以在这里执行重试
		if err != nil {
			h.l.Error("处理消息失败",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset))
		} else {
			session.MarkMessage(msg, "")
		}
	}

	return nil
}
